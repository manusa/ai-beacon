// Copyright 2026 Marc Nuri
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package huggingfacespace_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// RecipeSuite asserts the contract between the Hugging Face Spaces
// deployment recipe (Dockerfile + entrypoint + Space README.md) and the
// HF runtime. Failures here mean a Space built from this directory will
// not boot, will not authenticate, or will not connect agents.
type RecipeSuite struct {
	suite.Suite
	dockerfile string
	entrypoint string
	readme     string
}

func (s *RecipeSuite) SetupSuite() {
	dockerfile, err := os.ReadFile("Dockerfile")
	s.Require().NoError(err, "HF Space cannot build without a Dockerfile at the root of this directory")
	s.dockerfile = string(dockerfile)

	entrypoint, err := os.ReadFile("entrypoint.sh")
	s.Require().NoError(err, "HF Space container has no startup script — the image will exit immediately")
	s.entrypoint = string(entrypoint)

	readme, err := os.ReadFile("README.md")
	s.Require().NoError(err, "HF Space requires a README.md with YAML frontmatter declaring the Space configuration")
	s.readme = string(readme)
}

func (s *RecipeSuite) TestDockerfileFromPublishedImage() {
	s.Contains(s.dockerfile, "FROM ghcr.io/manusa/ai-beacon:latest",
		"recipe must wrap the published ai-beacon image at ghcr.io/manusa/ai-beacon:latest — drifting the base (e.g. an unpublished tag, a personal fork, the quay.io mirror that is no longer canonical) means the recipe ships an unverified binary")
}

func (s *RecipeSuite) TestDockerfileExposesAppPort() {
	s.Contains(s.dockerfile, "EXPOSE 8080",
		"HF Space app_port is 8080; without EXPOSE 8080 the Space will start but route to a closed port and time out")
}

// TestPortIsConsistentAcrossArtifacts pins the same port number in all
// three places it appears: Dockerfile EXPOSE, entrypoint.sh --address,
// and Space README.md app_port. A single edit in one of them produces
// a Space that builds and starts but routes to a closed port, which is
// slow to diagnose; this test fails fast on the drift.
func (s *RecipeSuite) TestPortIsConsistentAcrossArtifacts() {
	exposePort := firstMatch(regexp.MustCompile(`(?m)^EXPOSE\s+(\d+)`), s.dockerfile)
	s.Require().NotEmpty(exposePort, "Dockerfile must declare EXPOSE <port>")

	addressPort := firstMatch(regexp.MustCompile(`--address\s+"?:(\d+)"?`), s.entrypoint)
	s.Require().NotEmpty(addressPort, "entrypoint.sh must pass --address :<port> to ai-beacon")

	appPort := firstMatch(regexp.MustCompile(`(?m)^app_port:\s*(\d+)`), extractFrontmatter(s.readme))
	s.Require().NotEmpty(appPort, "Space README frontmatter must declare app_port: <port>")

	s.Equal(exposePort, addressPort,
		"Dockerfile EXPOSE %s and entrypoint.sh --address :%s disagree — HF Space will route to a port the binary isn't listening on", exposePort, addressPort)
	s.Equal(exposePort, appPort,
		"Dockerfile EXPOSE %s and Space README app_port: %s disagree — HF's proxy will forward to a port nothing binds", exposePort, appPort)
}

func (s *RecipeSuite) TestEntrypointPropagatesOIDCIssuer() {
	s.Contains(s.entrypoint, `--oidc-issuer "${OPENID_PROVIDER_URL}"`,
		"HF Space injects OPENID_PROVIDER_URL with the OIDC issuer URL; without this mapping ai-beacon starts without an issuer and OIDC sign-in is disabled")
}

func (s *RecipeSuite) TestEntrypointPropagatesOIDCClientID() {
	s.Contains(s.entrypoint, `--oidc-client-id "${OAUTH_CLIENT_ID}"`,
		"HF Space injects OAUTH_CLIENT_ID with the Space's OAuth client; without this mapping ai-beacon cannot identify itself to HF's IdP and every sign-in returns 401")
}

// TestEntrypointDoesNotExposeOIDCClientSecretOnArgv guards the recipe
// against materializing the OAuth client secret in /proc/<pid>/cmdline.
// The ai-beacon binary resolves OAUTH_CLIENT_SECRET from the
// environment (see ai-beacon/docs/auth.md § OIDC, and the
// auth.ProviderOIDC env-var fallback chain in pkg/cmd). Passing it as
// --oidc-client-secret would contradict the spec's stated preference
// for --oidc-client-secret-file precisely because argv is visible to
// anything that can read /proc.
func (s *RecipeSuite) TestEntrypointDoesNotExposeOIDCClientSecretOnArgv() {
	s.NotContains(s.entrypoint, "--oidc-client-secret ",
		"recipe must not pass the OAuth client secret on argv; let ai-beacon's env-var fallback read OAUTH_CLIENT_SECRET from the process environment instead")
	s.NotContains(s.entrypoint, "--oidc-client-secret=",
		"recipe must not pass the OAuth client secret on argv; let ai-beacon's env-var fallback read OAUTH_CLIENT_SECRET from the process environment instead")
}

func (s *RecipeSuite) TestEntrypointPropagatesOIDCScopes() {
	s.Contains(s.entrypoint, `--oidc-scopes "${OAUTH_SCOPES}"`,
		"HF Space injects OAUTH_SCOPES with the scopes the OAuth app was registered for; without this mapping ai-beacon may request scopes the IdP refuses and the IdP returns an authorization error")
}

func (s *RecipeSuite) TestEntrypointSetsAuthModeOIDC() {
	s.Contains(s.entrypoint, "--auth=oidc",
		"recipe must select OIDC browser auth; without --auth=oidc the Space falls back to the default password auth, which is incompatible with HF's OAuth gating and leaves the dashboard accessible to anyone with the password file")
}

func (s *RecipeSuite) TestEntrypointBuildsRedirectURLFromSpaceHost() {
	s.Contains(s.entrypoint, `--oidc-redirect-url "https://${SPACE_HOST}/login/oidc/callback"`,
		"HF Space hostname is injected as SPACE_HOST (no scheme); the recipe must derive the OIDC redirect URL from it — a hard-coded URL would only work for one specific Space")
}

func (s *RecipeSuite) TestSpaceREADMEDeclaresDockerSDK() {
	frontmatter := extractFrontmatter(s.readme)
	s.Require().NotEmpty(frontmatter, "HF Space README.md must open with a YAML frontmatter block — without it HF refuses to build the Space")
	s.Contains(frontmatter, "sdk: docker",
		"HF Space without sdk: docker would try to interpret the directory as a Gradio/Streamlit/Static Space and fail to build")
}

func (s *RecipeSuite) TestSpaceREADMEDeclaresAppPort() {
	frontmatter := extractFrontmatter(s.readme)
	s.Contains(frontmatter, "app_port: 8080",
		"HF Space defaults app_port to 7860; without app_port: 8080 the Space starts ai-beacon but HF routes external requests to the wrong port")
}

func (s *RecipeSuite) TestSpaceREADMEEnablesHFOAuth() {
	frontmatter := extractFrontmatter(s.readme)
	s.Contains(frontmatter, "hf_oauth: true",
		"without hf_oauth: true HF does not provision an OAuth client for the Space and the four env vars (OPENID_PROVIDER_URL, OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_SCOPES) are never injected — OIDC sign-in fails at boot")
}

// TestSpaceREADMEDeclaresOIDCScopes pins the OIDC scope contract.
//
// At runtime ai-beacon needs openid + profile + email so it can resolve
// preferred_username / email from the ID token. HF's hf_oauth_scopes
// YAML key is a strict allowlist of *additional* permission scopes
// (email, repo scopes, billing, …); openid and profile are auto-
// included by HF whenever hf_oauth: true is set, and listing them
// explicitly causes HF to reject the push with
//
//	Error: "hf_oauth_scopes[0]" must be one of [email, read-repos, …]
//
// So the recipe lists only `email`, and HF assembles
// `OAUTH_SCOPES="openid profile email"` at container start.
func (s *RecipeSuite) TestSpaceREADMEDeclaresOIDCScopes() {
	frontmatter := extractFrontmatter(s.readme)
	s.Contains(frontmatter, "- email",
		"HF Space must request the email scope so ai-beacon can resolve the signed-in user's email claim; without it the OIDC sign-in still works but the dashboard cannot show or allowlist users by email")
	for _, forbidden := range []string{"- openid", "- profile"} {
		s.NotContains(frontmatter, forbidden,
			"%q must not appear under hf_oauth_scopes — HF auto-includes openid and profile whenever hf_oauth: true is set, and listing them explicitly causes HF to reject the git push with 'must be one of [email, read-repos, ...]'", forbidden)
	}
}

// TestNoHFSpecificSymbolsInAuth guards the AC: the recipe must be pure
// config + image wrapper. No HF-specific auth wiring may leak into
// pkg/auth/. The display-name heuristic (oidc.go's switch over the
// issuer host that maps "huggingface" → "Hugging Face", parallel to
// "okta" → "Okta") is intentionally allowed — it's a generic IdP-label
// table, not an HF code path.
func (s *RecipeSuite) TestNoHFSpecificSymbolsInAuth() {
	authDir := filepath.Join("..", "pkg", "auth")

	// Tokens that only make sense for HF Spaces — none of them should
	// appear in pkg/auth/. Their presence would mean an HF-specific
	// branch has been smuggled into the generic OIDC code, breaking
	// the "pure config + image wrapper" guarantee of this issue.
	hfTokens := []string{
		"SPACE_HOST",     // HF-injected hostname env var — recipe-only concern
		"hf_oauth",       // HF Space metadata key — must never reach the auth code
		"hf.space",       // HF-hosted domain — should not be hardcoded anywhere in auth
		"huggingface.co", // HF-hosted issuer — should be config-driven, never hardcoded
	}

	err := filepath.WalkDir(authDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, token := range hfTokens {
			s.NotContains(string(content), token,
				"%s: HF-specific token %q leaked into pkg/auth/ — the HF Spaces recipe must remain a pure config + image wrapper that exercises the generic OIDC flags, not an HF-aware code path", path, token)
		}
		return nil
	})
	s.Require().NoError(err)
}

func TestRecipe(t *testing.T) {
	suite.Run(t, new(RecipeSuite))
}

// firstMatch returns the first capture group of the first regex match
// in s, or the empty string when there is no match. Used by the port-
// consistency assertion to read a single number out of each artifact.
func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// extractFrontmatter returns the contents of the leading "---\n...\n---"
// YAML block in an HF Space README. Returns the empty string when no
// frontmatter is present.
func extractFrontmatter(readme string) string {
	const sep = "---"
	if !strings.HasPrefix(readme, sep) {
		return ""
	}
	rest := readme[len(sep):]
	rest = strings.TrimPrefix(rest, "\n")
	idx := strings.Index(rest, "\n"+sep)
	if idx < 0 {
		return ""
	}
	return rest[:idx]
}
