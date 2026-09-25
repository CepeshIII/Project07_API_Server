# Changelog

## [1.3.0](https://github.com/CepeshIII/Project07_API_Server/compare/v1.2.0...v1.3.0) (2026-09-25)


### Features

* **cloudbuild:** add echo step and parameterize image paths in cloudbuild.yaml ([1971b8c](https://github.com/CepeshIII/Project07_API_Server/commit/1971b8c9d80dcd1ecabf27a4e3499e486c9e8a55))
* **cloudbuild:** add logging steps to cloudbuild.yaml for better visibility during builds ([23ca7e8](https://github.com/CepeshIII/Project07_API_Server/commit/23ca7e8c9bb6373b89ee5d63b5fdde337aaa89f8))
* dockerfile ([1a3231a](https://github.com/CepeshIII/Project07_API_Server/commit/1a3231afbdcdac6b46843a0f71ef7bfddeb079a7))
* **docker:** restructure Dockerfile for multi-stage builds and add development target; update entrypoint script and nginx configuration ([1c39173](https://github.com/CepeshIII/Project07_API_Server/commit/1c39173971e883f08802f378cd3c56b4a1e84a4f))
* implement multi-stage Dockerfile for frontend and backend, add entrypoint script, and configure Nginx ([e4f849d](https://github.com/CepeshIII/Project07_API_Server/commit/e4f849d0c48e98f7e66f1f2a720b77c6003ff944))


### Bug Fixes

* **cloudbuild:** add DOCKER_BUILDKIT environment variable for build step ([0743122](https://github.com/CepeshIII/Project07_API_Server/commit/0743122eed7ae0ce1a1037d68f291cd4d8ad577e))
* **cloudbuild:** change docker build target from 'build' to 'production' ([0ceabbe](https://github.com/CepeshIII/Project07_API_Server/commit/0ceabbe604ac5f0b8066678f45269c21f1b02e23))
* **cloudbuild:** remove empty string argument from docker build step ([dcb8c01](https://github.com/CepeshIII/Project07_API_Server/commit/dcb8c0150aebbb21eb4b3a40705f709b7a396837))
* **docker-compose:** set DOCKER_BUILDKIT environment variable for fullstack service ([0743122](https://github.com/CepeshIII/Project07_API_Server/commit/0743122eed7ae0ce1a1037d68f291cd4d8ad577e))
* **dockerfile:** remove commented cache mount commands for clarity ([e4eba55](https://github.com/CepeshIII/Project07_API_Server/commit/e4eba5517219a015a1d40f270b95fb008cf63711))
* **docker:** make production stage the default target ([634f4f0](https://github.com/CepeshIII/Project07_API_Server/commit/634f4f018a78d0a3246858ad6ee1c10804834317))
* **GetUserPage:** correct API endpoint URL for fetching user data ([7f0a698](https://github.com/CepeshIII/Project07_API_Server/commit/7f0a6989174daf4b6663ad38b3725eebd41f8daa))
* **GetUserPage:** update API endpoint for user data fetching; refactor API_URL in App component ([07471b1](https://github.com/CepeshIII/Project07_API_Server/commit/07471b10fedb3be5e35924d713a0dde0de0cd7d3))
* **server:** handle missing .env file gracefully and prevent nil Redis panic ([ecd3616](https://github.com/CepeshIII/Project07_API_Server/commit/ecd3616b7980446fa07392b70760cd1d749b891f))
* update database connection string and adjust Dockerfile build command ([e4dd748](https://github.com/CepeshIII/Project07_API_Server/commit/e4dd7487097eeb1e8d788f2b2c862e1fb0394df3))

## [1.1.0](https://github.com/CepeshIII/Project07_API_Server/compare/v1.0.0...v1.1.0) (2026-09-18)


### Features

* update api version automatically ([3f36e53](https://github.com/CepeshIII/Project07_API_Server/commit/3f36e530fc7a012cdc827f7e239ca5d95a0641ee))

## 1.0.0 (2026-09-17)


### Features

* add Docker start command to MakeFile and release please scripts to github workflows ([cf3a2c5](https://github.com/CepeshIII/Project07_API_Server/commit/cf3a2c56277d163ced7a601e43e2a09c0a298510))
* add release-please GitHub Actions workflow for automated releases ([2502ff2](https://github.com/CepeshIII/Project07_API_Server/commit/2502ff2a1f78b2045eb1293c02cce01d7efb4d5a))
* Add Swagger documentation for GopherSocial API ([a349086](https://github.com/CepeshIII/Project07_API_Server/commit/a3490864e4feedcbebab96204642fd31a5d96259))
* add user deletion endpoint and improve email sending logic ([5a560e1](https://github.com/CepeshIII/Project07_API_Server/commit/5a560e1988b8b779a3b0ba8fd705363186db44de))
* Add user invitation email template ([a349086](https://github.com/CepeshIII/Project07_API_Server/commit/a3490864e4feedcbebab96204642fd31a5d96259))
* Implement logger package ([a349086](https://github.com/CepeshIII/Project07_API_Server/commit/a3490864e4feedcbebab96204642fd31a5d96259))
* implement user follow/unfollow functionality and user feed retrieval ([c958430](https://github.com/CepeshIII/Project07_API_Server/commit/c958430c0d62cf20a60a165e11fa4dd270f8ff27))
* initialize web application with React and Vite ([d658df5](https://github.com/CepeshIII/Project07_API_Server/commit/d658df5e48bec1248ba9af93d80078fa401ba3d6))
* Introduce mailer package with SendGrid integration ([a349086](https://github.com/CepeshIII/Project07_API_Server/commit/a3490864e4feedcbebab96204642fd31a5d96259))


### Bug Fixes

* add workflow_dispatch trigger to audit.yaml ([7a96bf0](https://github.com/CepeshIII/Project07_API_Server/commit/7a96bf00ff10a7bbae0f02a648476470942ffdbb))
* audit.yaml ([f88b90b](https://github.com/CepeshIII/Project07_API_Server/commit/f88b90b1b0fcac4992ef0997ddfc18e5c6e50014))
* audit.yaml ([24a8f71](https://github.com/CepeshIII/Project07_API_Server/commit/24a8f710b7f166fed463e43f6e491499fca75316))
* handle error in Delete method and simplify password hash check ([fd4224f](https://github.com/CepeshIII/Project07_API_Server/commit/fd4224f3f1df3819089c6035c9d5cb8959d38bd1))
* standardize error messages for unauthorized access in auth.go ([61b82d6](https://github.com/CepeshIII/Project07_API_Server/commit/61b82d68ecc31418b5ca7984c73789d12307c088))
* standardize unauthorized error messages in auth and middleware ([a74f428](https://github.com/CepeshIII/Project07_API_Server/commit/a74f4286a5612e349fbc1a5db611e0507356a194))
* update audit.yaml for workflow_dispatch trigger and action versions ([f68c86c](https://github.com/CepeshIII/Project07_API_Server/commit/f68c86cc487c874b2f1dd1da95aefe16c11a15e3))
* update branch name from main to master in release-please workflow ([4e25c70](https://github.com/CepeshIII/Project07_API_Server/commit/4e25c707eb84fc3cd80978dbea44288ea83229cd))
* update branch references from main to master in audit.yaml ([9694b7b](https://github.com/CepeshIII/Project07_API_Server/commit/9694b7ba5590c96f27027aed081866f00f169b48))
