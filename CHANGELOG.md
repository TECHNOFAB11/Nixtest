# Changelog
All notable changes to this project will be documented in this file. See [conventional commits](https://www.conventionalcommits.org/) for commit guidelines.

- - -

## [v1.2.1](https://gitlab.com/TECHNOFAB/nixtest/compare/v1.2.0..v1.2.1) - 2026-04-02
#### Documentation
- set _file so "declared in" works correctly - ([5a7053a](https://gitlab.com/TECHNOFAB/nixtest/commit/5a7053afcbb211b9cf8fe87f7892bb9f6b76b678)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- update nixmkdocs, use svg for logo and favicon, add module docs - ([c9618a4](https://gitlab.com/TECHNOFAB/nixtest/commit/c9618a4d9b03939ff2673e52aad97b57acd0f45a)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Tests
- set SSL_CERT_FILE and NIX_SSL_CERT_FILE so test works in pure mode - ([56d22f4](https://gitlab.com/TECHNOFAB/nixtest/commit/56d22f4aa1c3f281a0cd26acd7aa2ed426ef6fe5)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Refactoring
- replace flake-parts, devenv etc. with rensa - ([0414493](https://gitlab.com/TECHNOFAB/nixtest/commit/041449396326a11ec8f88354f7dbd3204016de1d)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Miscellaneous Chores
- rename module to uppercase - ([318b903](https://gitlab.com/TECHNOFAB/nixtest/commit/318b903d122b7628181f2ae277a882e2328b1d1d)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)

- - -

## [v1.2.0](https://gitlab.com/TECHNOFAB/nixtest/compare/v1.1.0..v1.2.0) - 2026-04-02
#### Features
- run script tests in temp dirs for slightly better sandboxing - ([5741109](https://gitlab.com/TECHNOFAB/nixtest/commit/5741109cc9ec2b6d41b56abd3f5bc51ed7a9a228)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Bug Fixes
- (**README**) fix badge - ([d7e4902](https://gitlab.com/TECHNOFAB/nixtest/commit/d7e4902fede9e03073207a3f3f1ca34c9d0e1c70)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- support passing string as dir for autodiscovery - ([0272a8b](https://gitlab.com/TECHNOFAB/nixtest/commit/0272a8b0dc8f64c0b590f7007291b9010a0eef44)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Miscellaneous Chores
- (**flake**) enable go hardening workaround - ([22b43c9](https://gitlab.com/TECHNOFAB/nixtest/commit/22b43c9fe83be73c3f0648bbb54bc3c1cf7f96df)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- ![BREAKING](https://img.shields.io/badge/BREAKING-red) default to pure mode, rename --pure flag to --impure for switching - ([c9298b9](https://gitlab.com/TECHNOFAB/nixtest/commit/c9298b91f42a2f02842da6c41d34db342d4b3de6)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- clean up module a bit - ([b2fb77e](https://gitlab.com/TECHNOFAB/nixtest/commit/b2fb77ecc9d48556801d3bfba3b547b754f2aedc)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)

- - -

## [v1.1.0](https://gitlab.com/TECHNOFAB/nixtest/compare/v1.0.0..v1.1.0) - 2026-04-02
#### Features
- add test helpers - ([bed029f](https://gitlab.com/TECHNOFAB/nixtest/commit/bed029f4a90b886bb85aef76264e5af40d0d3938)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- switch to module system to evaluate suites & tests - ([98141a1](https://gitlab.com/TECHNOFAB/nixtest/commit/98141a1f5c0c1b48837cbc363cbb8ebf74b5d044)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Documentation
- document new lib functions & usage - ([bc36c39](https://gitlab.com/TECHNOFAB/nixtest/commit/bc36c39b0929bdfaab0908d8bd852c2114e3383f)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- update cli help - ([e8da91a](https://gitlab.com/TECHNOFAB/nixtest/commit/e8da91ad27f4e73d22d095c8b0ca607bd50880d9)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add images & fix typo - ([6ee3811](https://gitlab.com/TECHNOFAB/nixtest/commit/6ee3811b568f6e840a964a8851719638234cd03c)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Tests
- nest sample tests as fixtures in real tests - ([006537e](https://gitlab.com/TECHNOFAB/nixtest/commit/006537e6abbcb6d6b78dd1c76e1c937b01378abd)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Continuous Integration
- remove flakeModule test job - ([3f1b631](https://gitlab.com/TECHNOFAB/nixtest/commit/3f1b6317b4c43ab0e17c194f02c6a8ec138360fe)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Miscellaneous Chores
- (**cli**) handle help command manually to exit 0 - ([4a8ccdf](https://gitlab.com/TECHNOFAB/nixtest/commit/4a8ccdf34cff2c91bab5fa906f6ed4e6b0e928de)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- (**module**) remove "set -x" addition in script tests - ([001b575](https://gitlab.com/TECHNOFAB/nixtest/commit/001b575f31f45409b953371a56509ed3f2035f08)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- (**testHelpers**) remove string context in toJsonFile - ([116f905](https://gitlab.com/TECHNOFAB/nixtest/commit/116f905b6c1df01f7c2dddef4f662abd0fcb5d42)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add more helpers & add pos to test suite - ([4a55db9](https://gitlab.com/TECHNOFAB/nixtest/commit/4a55db97979b8b06ebd68b4b0772fcb76d79d2fa)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- merge branch 'main' into feat/module-system - ([3bb1764](https://gitlab.com/TECHNOFAB/nixtest/commit/3bb1764013f37da0667d12a565a67abf76546e49)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add LICENSE - ([b732e11](https://gitlab.com/TECHNOFAB/nixtest/commit/b732e118dfcd8b2f9b8b2302ee96ad0644f9923f)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)

- - -

## [v1.0.0](https://gitlab.com/TECHNOFAB/nixtest/compare/c1c19c324d67dbcd482ed9cd3dac05f74a042a98..v1.0.0) - 2026-04-02
#### Features
- add pure mode which unsets env variables before script tests - ([3ff5b35](https://gitlab.com/TECHNOFAB/nixtest/commit/3ff5b358d506473ace2311cd25fb95be0c048997)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- sort suites and tests alphabetically in the console summary - ([7bb5c65](https://gitlab.com/TECHNOFAB/nixtest/commit/7bb5c65259ab606de337959c308f9de3b7eefbc0)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add "script" test type - ([b50a1c6](https://gitlab.com/TECHNOFAB/nixtest/commit/b50a1c61a6b4338bbaed4da2dcdf8eadc7c56f84)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add support for pretty/nix format - ([e029fae](https://gitlab.com/TECHNOFAB/nixtest/commit/e029fae0b86b8b6ee39368f7870585809c050b8f)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- general improvements and add junit "error" and "skipped" support - ([0a1bbae](https://gitlab.com/TECHNOFAB/nixtest/commit/0a1bbae2c30e3ba8e3b02de223e199c5dfd56572)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- remove nix store path prefix if used in flakeModule - ([482f15c](https://gitlab.com/TECHNOFAB/nixtest/commit/482f15c486380455c909b309b94e0b26b5faa362)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Bug Fixes
- (**flake**) add missing lib - ([0b78315](https://gitlab.com/TECHNOFAB/nixtest/commit/0b783157bb59a2f9db4a9f5d810be2fdfe37fb15)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- (**lib**) discard string context so derivations are not built instantly - ([5ae5c2d](https://gitlab.com/TECHNOFAB/nixtest/commit/5ae5c2dd4508f4476a0c50fb6668af8fde151cb6)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- (**package**) use filesets so nixtest doesnt get rebuilt all the time - ([25de506](https://gitlab.com/TECHNOFAB/nixtest/commit/25de5061ad1b53b03cf02ddd13deb3ce920e0545)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- handle final newline in messages correctly - ([5436abf](https://gitlab.com/TECHNOFAB/nixtest/commit/5436abf377ce739958f5fa74e71004bf78a0c744)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- display multi-line diffs with correct colors - ([2b7c215](https://gitlab.com/TECHNOFAB/nixtest/commit/2b7c215ffa8568f04ccfcd8995a21428a158cdb5)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- handle error instead of including it in diff - ([d0e47e3](https://gitlab.com/TECHNOFAB/nixtest/commit/d0e47e305dc637ea4cf0565dfab3cc115d597377)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- pretty print expected aswell when format is "pretty" - ([9068477](https://gitlab.com/TECHNOFAB/nixtest/commit/90684776508b6f7270892b83959ccfc84fa86649)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- handle nix build errors gracefully - ([6e17ec8](https://gitlab.com/TECHNOFAB/nixtest/commit/6e17ec8838ef1768e1d4aeebc4c79a2094c5cc7b)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- allow snapshots to use actualDrv aswell - ([abac8aa](https://gitlab.com/TECHNOFAB/nixtest/commit/abac8aaf3e978e02e309235eeb8bd9443d561b18)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Documentation
- add documentation - ([fd58344](https://gitlab.com/TECHNOFAB/nixtest/commit/fd58344ca7173f7d66a19489657e1c21bb00efa9)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Tests
- add drv test - ([27696c0](https://gitlab.com/TECHNOFAB/nixtest/commit/27696c02bc17b7ebe7c6c7eae15756e3ce976ee2)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Refactoring
- split into packages and add tests - ([11117e0](https://gitlab.com/TECHNOFAB/nixtest/commit/11117e0c0ef23963e9a385f9cb424ba5f0f04884)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- improve junit report to not include escape sequences - ([3a974f2](https://gitlab.com/TECHNOFAB/nixtest/commit/3a974f218afddff5132e61524d14613d964a90f7)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- rename variable to derivation to be more descriptive - ([4afa8a7](https://gitlab.com/TECHNOFAB/nixtest/commit/4afa8a7957ef5c4cc1d928c46896288fdb9f3931)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
#### Miscellaneous Chores
- (**lib**) add assertion for script to not be null when type=="script" - ([c2ca17d](https://gitlab.com/TECHNOFAB/nixtest/commit/c2ca17dfc514228b86a7d503c65c2e384d02b671)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- use set -x to show which line of "script" failed the test - ([f59d927](https://gitlab.com/TECHNOFAB/nixtest/commit/f59d92791bea66b0e969e96d8a702cb841f1f1b0)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- display "script" output on failure without diffing - ([c1ad61b](https://gitlab.com/TECHNOFAB/nixtest/commit/c1ad61b3674b9b258deb4fab2d2dc3820f5ffde5)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- use git-like patch/diff in junit output - ([8b5ac90](https://gitlab.com/TECHNOFAB/nixtest/commit/8b5ac904dbf2a3c6d08b8d7b8973b9d3fe4ad777)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- allow setting pos for entire suite & remove column from pos - ([4aaaf32](https://gitlab.com/TECHNOFAB/nixtest/commit/4aaaf32621b644d9e1f8127a9c887aa0a6b1a7e6)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- add CI to dogfood - ([4772c78](https://gitlab.com/TECHNOFAB/nixtest/commit/4772c789d9f210fe9be720fe284fce18f18027bc)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)
- initial prototype - ([c1c19c3](https://gitlab.com/TECHNOFAB/nixtest/commit/c1c19c324d67dbcd482ed9cd3dac05f74a042a98)) - [@TECHNOFAB](https://gitlab.com/TECHNOFAB)

- - -

Changelog generated by [cocogitto](https://github.com/cocogitto/cocogitto).


