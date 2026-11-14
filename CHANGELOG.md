# Changelog

## [0.17.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.16.0...ponto_certo-v0.17.0) (2026-11-14)


### Features

* **api:** add owner commissions multi-level update endpoints ([2d42256](https://github.com/matheuseufrasio2/ponto_certo/commit/2d42256569d8fd3c8c6b88e4c71b69bf022904c0))
* **api:** add owner commissions multi-level update endpoints ([7541829](https://github.com/matheuseufrasio2/ponto_certo/commit/75418298861de876e127cfd5b79c01eb87bd8704))

## [0.16.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.15.0...ponto_certo-v0.16.0) (2026-11-14)


### Features

* **api:** add owner unit managers endpoints ([37d6adf](https://github.com/matheuseufrasio2/ponto_certo/commit/37d6adf59829584c2025f7f2130025516a43eae8))
* **api:** add owner unit managers endpoints ([c55e838](https://github.com/matheuseufrasio2/ponto_certo/commit/c55e8380fd24ddf1a3ea72be5fb354ac33f19037))

## [0.15.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.14.0...ponto_certo-v0.15.0) (2026-11-12)


### Features

* **owner:** implement branch unit management endpoints ([7ff5495](https://github.com/matheuseufrasio2/ponto_certo/commit/7ff5495edb6270cc368c50a5b85306a293f52012))


### Bug Fixes

* **repo:** guard invalid uuid in IsUnitActive ([6ce01fd](https://github.com/matheuseufrasio2/ponto_certo/commit/6ce01fdd7c9227a8cf1855f3f3e723925d2642de))

## [0.14.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.13.0...ponto_certo-v0.14.0) (2026-11-11)


### Features

* **owner:** implement organization profile endpoints ([f349b25](https://github.com/matheuseufrasio2/ponto_certo/commit/f349b25295b9e21129c923bf382e26158ae4cf78))
* **owner:** implement organization profile endpoints ([6f35c6e](https://github.com/matheuseufrasio2/ponto_certo/commit/6f35c6ec7a2ce8a3ab801cc04cdcea7ca2b95833))

## [0.13.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.12.0...ponto_certo-v0.13.0) (2026-11-08)


### Features

* **domain:** define canonical error codes ([85b9ea0](https://github.com/matheuseufrasio2/ponto_certo/commit/85b9ea08ac3fa7ffd7842b367fe8fb57fe6510ae))
* **handler:** add client message localization support ([0eb6844](https://github.com/matheuseufrasio2/ponto_certo/commit/0eb68445db1137fa59e71c5f8db06a2c06c74922))
* **i18n:** implement multilingual error localization with embedded json ([f2f5753](https://github.com/matheuseufrasio2/ponto_certo/commit/f2f5753bfa8921d1244c705680e196b6df32c1ca))
* **i18n:** implement multilingual error localization with embedded json ([39d6c46](https://github.com/matheuseufrasio2/ponto_certo/commit/39d6c46a9334eefff6c1864b811221af7938e581))

## [0.12.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.11.0...ponto_certo-v0.12.0) (2026-11-08)


### Features

* **domain:** define governance audit logs with employee leave entities ([f714fdc](https://github.com/matheuseufrasio2/ponto_certo/commit/f714fdcb25aa4f581fcf6e5bf53d456b9001c639))
* **governance:** implement system audit logs with team leave governance ([5b1e1b3](https://github.com/matheuseufrasio2/ponto_certo/commit/5b1e1b38911ef16ca531fdb306265c7cfcfb594a))
* **handler:** add endpoints for user governance with audit logs ([febca0e](https://github.com/matheuseufrasio2/ponto_certo/commit/febca0e69cdba2389c9ee007aec9ca8dec818bb0))
* **migrations:** add audit logs with employee leaves schema ([b973574](https://github.com/matheuseufrasio2/ponto_certo/commit/b97357472b8febf4b2dc03f99ee36bc19ac80bb4))
* **repo:** implement audit logs persistence with multi-tenant filtering ([c1686ac](https://github.com/matheuseufrasio2/ponto_certo/commit/c1686aca834565583813d9b779086a9f33881168))
* **service:** implement governance services with audit trail recording ([89a73e7](https://github.com/matheuseufrasio2/ponto_certo/commit/89a73e7db120ceda89297dd1f922fb114dde8cea))

## [0.11.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.10.0...ponto_certo-v0.11.0) (2026-11-08)


### Features

* **admin:** implement phase 4 saas superadmin platform ([5ced9e2](https://github.com/matheuseufrasio2/ponto_certo/commit/5ced9e23bbe07048a2c46e19ae084a57bec50928))
* **api:** register dev routes with tenant status protection ([5d23594](https://github.com/matheuseufrasio2/ponto_certo/commit/5d235941d8ffd9dd66a38812f91af6b827d9c67e))
* **domain:** add subscription models with dev discovery structures ([ccbcc6a](https://github.com/matheuseufrasio2/ponto_certo/commit/ccbcc6a4e2d828896d549cdfc32c4b34ece4ce79))
* **handler:** add dev discovery endpoints ([82954d2](https://github.com/matheuseufrasio2/ponto_certo/commit/82954d25fba9ab74f048dbd7d383554c80ae5a73))
* **handler:** add owner subscription endpoint ([ac9ce89](https://github.com/matheuseufrasio2/ponto_certo/commit/ac9ce89ee7a33d353b00dfd08efb5adba2f1a874))
* **middleware:** enforce active tenant check on operational routes ([481b7d6](https://github.com/matheuseufrasio2/ponto_certo/commit/481b7d6cbe6627f9e10f333042697bc48f62ebd1))
* **repo:** add organization active status queries ([f17a4e1](https://github.com/matheuseufrasio2/ponto_certo/commit/f17a4e1d91181e123e907b1193afe8b6665db3d3))
* **repo:** implement developer discovery queries ([4f24fa9](https://github.com/matheuseufrasio2/ponto_certo/commit/4f24fa96d61882a9cc2e5d95ab759da143552779))
* **repo:** implement owner subscription query ([656b7ab](https://github.com/matheuseufrasio2/ponto_certo/commit/656b7abdbb4df27d21d7e720bcef51ab11827657))
* **service:** block non-owner logins for inactive organizations ([e976cfc](https://github.com/matheuseufrasio2/ponto_certo/commit/e976cfc85b5622fa35bd2e15eca68afc815ab5de))
* **service:** implement developer discovery service ([79929ca](https://github.com/matheuseufrasio2/ponto_certo/commit/79929ca490b663abf95ce1cb58a4afac80efe2bf))
* **service:** implement owner subscription retrieval ([87a35d2](https://github.com/matheuseufrasio2/ponto_certo/commit/87a35d28ff4cd45b1037677c35f7d402c846bf52))

## [0.10.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.9.0...ponto_certo-v0.10.0) (2026-11-08)


### Features

* **api:** implement tenant units endpoint with status plan filtering ([fc2f941](https://github.com/matheuseufrasio2/ponto_certo/commit/fc2f9417c1f94b64d38979502213cb073c44de15))
* **api:** register tenant overview plan with password reset endpoints ([6a68ccd](https://github.com/matheuseufrasio2/ponto_certo/commit/6a68ccd89f99c32f1738b44962fcd110b94c6c0c))
* **api:** register tenant units endpoint with filter params ([3bf325c](https://github.com/matheuseufrasio2/ponto_certo/commit/3bf325cbbdd685c02a3e9005d84bf2ca9aace206))
* **cli:** add database seed runner with 45 tenants ([ce843d1](https://github.com/matheuseufrasio2/ponto_certo/commit/ce843d13abf42b1ce1bfdca50273978a74f9ef8d))
* **cli:** introduce native go database migration runner ([9ba2ac5](https://github.com/matheuseufrasio2/ponto_certo/commit/9ba2ac5000d141a942cfa1c9b141effded4bd2ff))
* **domain:** add tenant operational counters with overview structures ([cb1c642](https://github.com/matheuseufrasio2/ponto_certo/commit/cb1c6427c81e4461b3762fe0b3ffd9f9eb5e68fa))
* **handler:** validate uuid format for tenant admin endpoints ([ce5a8b0](https://github.com/matheuseufrasio2/ponto_certo/commit/ce5a8b0d7bb0f8137f386741828433e5c6b54d02))
* **migrations:** add past due since timestamp to organizations ([f109eed](https://github.com/matheuseufrasio2/ponto_certo/commit/f109eed63a60fd2ba148136b4070049262742974))
* **repository:** implement tenant overview with plan update queries ([28e129c](https://github.com/matheuseufrasio2/ponto_certo/commit/28e129c4ba2f5d5c62ec10492da8fb1b5d0917ed))
* **repository:** implement tenant units query with status plan filtering ([6d77aec](https://github.com/matheuseufrasio2/ponto_certo/commit/6d77aecb3cee17ea57e5b7e17c7602560be1a299))
* **service:** implement tenant overview with administrative password reset ([4503131](https://github.com/matheuseufrasio2/ponto_certo/commit/45031316e554632d0d61f6a6fe19a87f6a920d47))
* **service:** implement tenant units retrieval with filter forwarding ([8b9085a](https://github.com/matheuseufrasio2/ponto_certo/commit/8b9085ab48e5ae6e20c554917f08e5cb62d0771d))
* **service:** update default tenant pagination limit to 10 ([3d2a5bd](https://github.com/matheuseufrasio2/ponto_certo/commit/3d2a5bd5b1c9a496481eabd2e42ee8b7ca7514d5))

## [0.9.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.8.0...ponto_certo-v0.9.0) (2026-11-07)


### Features

* **admin:** implement saas governance and diagnostic user impersonation ([8dc3aa8](https://github.com/matheuseufrasio2/ponto_certo/commit/8dc3aa8bbd688b0c3ea2f8a6b3a98eed32603bff))
* **api:** register saas platform admin endpoints ([451d5f9](https://github.com/matheuseufrasio2/ponto_certo/commit/451d5f95d6063dc3e2b95d23d7eda47921384f82))
* **db:** add saas governance columns to organizations ([451b82c](https://github.com/matheuseufrasio2/ponto_certo/commit/451b82c707eb311809756e1eee10f608efee1900))
* **domain:** add saas governance impersonation models ([1265d87](https://github.com/matheuseufrasio2/ponto_certo/commit/1265d8743d655bda43296990a987d75ada206fab))
* **middleware:** add impersonation mutation guard ([30162b1](https://github.com/matheuseufrasio2/ponto_certo/commit/30162b1a1b077dea12951b6f8b42561fa66ae21f))
* **repository:** implement tenant admin repository ([fc959ab](https://github.com/matheuseufrasio2/ponto_certo/commit/fc959ab552522b6da03810914561e50ee8660bb1))
* **service:** implement tenant admin impersonation logic ([548b265](https://github.com/matheuseufrasio2/ponto_certo/commit/548b265cbac7e5c604ef4ee0f68f6300bc2c9ba2))

## [0.8.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.7.0...ponto_certo-v0.8.0) (2026-11-07)


### Features

* **api:** register owner analytics endpoints ([7407c6f](https://github.com/matheuseufrasio2/ponto_certo/commit/7407c6fe61696eb8c7085c2322ddd2a20264da2c))
* **domain:** add owner bi executive structures ([6243473](https://github.com/matheuseufrasio2/ponto_certo/commit/6243473d2078d0307a0043075409bb5c2029df95))
* **owner:** implement executive network bi and financial analytics for phase 3 ([6eb7407](https://github.com/matheuseufrasio2/ponto_certo/commit/6eb740782da45cb62cf1ae955911c0eed234cce6))
* **repository:** implement owner analytics repository ([ea312ed](https://github.com/matheuseufrasio2/ponto_certo/commit/ea312edbea4f19604eab796b5a95559c66be48fd))
* **service:** implement owner analytics service ([29d2e8b](https://github.com/matheuseufrasio2/ponto_certo/commit/29d2e8b27e83dd1a2591db51c92cbe53fd318442))

## [0.7.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.6.0...ponto_certo-v0.7.0) (2026-11-07)


### Features

* **manager:** implement unit operational management endpoints for phase 2 ([5860d92](https://github.com/matheuseufrasio2/ponto_certo/commit/5860d92c68663e5986910310db37e63bea2b2709))


### Bug Fixes

* **repository:** handle rollback return value in unit operations ([248555f](https://github.com/matheuseufrasio2/ponto_certo/commit/248555f0c108e42917bcf69e2dbae231f579d0d7))

## [0.6.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.5.0...ponto_certo-v0.6.0) (2026-11-07)


### Features

* **logger:** implement contextual request logging middleware ([065615f](https://github.com/matheuseufrasio2/ponto_certo/commit/065615f7618f98c91ff9ec0536074a49e72b7377))
* **seed:** add second organization for multi-tenant testing ([f729a96](https://github.com/matheuseufrasio2/ponto_certo/commit/f729a967eac2f133ce92e99d6671043da41f75a6))

## [0.5.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.4.0...ponto_certo-v0.5.0) (2026-11-06)


### Features

* **reception:** implement operational endpoints ([88cdafc](https://github.com/matheuseufrasio2/ponto_certo/commit/88cdafceb18534878d4f008d1eeb03e1898630f2))
* **reception:** implement phase 1 operational endpoints ([91c2711](https://github.com/matheuseufrasio2/ponto_certo/commit/91c2711b698c5774a108433f4e3f5d4953280ec9))


### Bug Fixes

* **billing:** ignore unchecked transaction rollback error in defer ([b3de8a2](https://github.com/matheuseufrasio2/ponto_certo/commit/b3de8a2fd42f3ce3bdf90cef83024f94b1043a39))

## [0.4.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.3.0...ponto_certo-v0.4.0) (2026-11-06)


### Features

* add appointments, retail products, and realistic database seed ([2d6c648](https://github.com/matheuseufrasio2/ponto_certo/commit/2d6c648b1c3666c0080a6760cb4f0565a5960b8d))
* **db:** add appointments, products, schedules schema ([fc21c3a](https://github.com/matheuseufrasio2/ponto_certo/commit/fc21c3a9c104a78f27bc591d89fe00980f025a3b))
* **db:** create realistic database seed with products, make target ([f5f67cd](https://github.com/matheuseufrasio2/ponto_certo/commit/f5f67cd8bac163758587bcc4115381c11bbe5e68))
* **domain:** add appointment, product domain models with validations ([b836428](https://github.com/matheuseufrasio2/ponto_certo/commit/b8364280db002f0c0bbe6b333c663a31f5cbd1a9))

## [0.3.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.2.0...ponto_certo-v0.3.0) (2026-11-06)


### Features

* implement multitenant database schema, rbac domain models, and jwt authentication ([#3](https://github.com/matheuseufrasio2/ponto_certo/issues/3)) ([e0c5427](https://github.com/matheuseufrasio2/ponto_certo/commit/e0c542706534015442696d0796b0ea398933ddc4))

## [0.2.0](https://github.com/matheuseufrasio2/ponto_certo/compare/ponto_certo-v0.1.0...ponto_certo-v0.2.0) (2026-10-31)


### Features

* **domain:** define client, billing models and domain errors ([a415365](https://github.com/matheuseufrasio2/ponto_certo/commit/a415365b2d462dcd8b63d5e7d767682817b6555c))
* initialize barbershop api with clean architecture, ci and docs ([0bb7cb0](https://github.com/matheuseufrasio2/ponto_certo/commit/0bb7cb088bb08756597c9540c1d3979c35017886))
* **migrations:** add initial schema for barbers, clients, services and billings ([053be44](https://github.com/matheuseufrasio2/ponto_certo/commit/053be44f06c07aa1a94439e1c839e41b16afa8f4))
* **repo:** configure postgresql docker-compose and pgxpool connection ([ae81986](https://github.com/matheuseufrasio2/ponto_certo/commit/ae819869568c9b20dbc398b338a3b4e1538daa08))
