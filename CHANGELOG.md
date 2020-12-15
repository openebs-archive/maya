v2.4.0 / 2020-12-13
========================
* fix(jiva): add serviceAccount name in jiva deployments ([#1773](https://github.com/openebs/maya/pull/1773),[@prateekpandey14](https://github.com/prateekpandey14))
* chore(version): add 2.4.0 to upgrade matrix ([#1772](https://github.com/openebs/maya/pull/1772),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fix(webhook): update webhook config sideEffects to None ([#1774](https://github.com/openebs/maya/pull/1774),[@prateekpandey14](https://github.com/prateekpandey14))


v2.4.0-RC2 / 2020-12-12
========================


v2.4.0-RC1 / 2020-12-10
========================
* fix(jiva): add serviceAccount name in jiva deployments ([#1773](https://github.com/openebs/maya/pull/1773),[@prateekpandey14](https://github.com/prateekpandey14))
* chore(version): add 2.4.0 to upgrade matrix ([#1772](https://github.com/openebs/maya/pull/1772),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fix(webhook): update webhook config sideEffects to None ([#1774](https://github.com/openebs/maya/pull/1774),[@prateekpandey14](https://github.com/prateekpandey14))


v2.3.0 / 2020-11-14
========================
* fix(restore): set targetip on CVRs after restore is completed ([#1761](https://github.com/openebs/maya/pull/1761),[@zlymeda](https://github.com/zlymeda))
* refactor(localpv): move builds to dynamic-localpv-provisioner repo ([#1762](https://github.com/openebs/maya/pull/1762),[@kmova](https://github.com/kmova))
* fix(upgrade): remove quay.io as default url-prefix ([#1768](https://github.com/openebs/maya/pull/1768),[@shubham14bajpai](https://github.com/shubham14bajpai))
* chore(build): add support for multi-arch builds ([#1764](https://github.com/openebs/maya/pull/1764),[@shubham14bajpai](https://github.com/shubham14bajpai))
* refactor(exporter): move builds to openebs-exporter repo ([#1763](https://github.com/openebs/maya/pull/1763),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fix(upgrade): allow upgrades for volumes without monitor ([#1760](https://github.com/openebs/maya/pull/1760),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.3.0-RC2 / 2020-11-13
========================
* No changes

v2.3.0-RC1 / 2020-11-13
========================
* fix(restore): set targetip on CVRs after restore is completed ([#1761](https://github.com/openebs/maya/pull/1761),[@zlymeda](https://github.com/zlymeda))
* refactor(localpv): move builds to dynamic-localpv-provisioner repo ([#1762](https://github.com/openebs/maya/pull/1762),[@kmova](https://github.com/kmova))
* fix(upgrade): remove quay.io as default url-prefix ([#1768](https://github.com/openebs/maya/pull/1768),[@shubham14bajpai](https://github.com/shubham14bajpai))
* chore(build): add support for multi-arch builds ([#1764](https://github.com/openebs/maya/pull/1764),[@shubham14bajpai](https://github.com/shubham14bajpai))
* refactor(exporter): move builds to openebs-exporter repo ([#1763](https://github.com/openebs/maya/pull/1763),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fix(upgrade): allow upgrades for volumes without monitor ([#1760](https://github.com/openebs/maya/pull/1760),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.2.0-RC2 / 2020-10-14
========================

 * No changes

v2.2.0-RC2 / 2020-10-12
========================

 * No changes

v2.2.0-RC1 / 2020-10-08
========================

 * No changes

v2.1.0 / 2020-09-14
========================

 * feat(webhook): add validation for namspace delete requests [#1757](https://github.com/openebs/maya/pull/1757) ([shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(upgrade): add valid current versions for 2.1.0 upgrades [#1749](https://github.com/openebs/maya/pull/1749) ([shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(spc): add capability to specify allowed BD tags on SPC [#1748](https://github.com/openebs/maya/pull/1748) ([sonasingh46](https://github.com/sonasingh46))
 * fix(spc): provision cStor stripe based pools with single raid group [#1744](https://github.com/openebs/maya/pull/1744) ([mittachaitu](https://github.com/mittachaitu))
 * fix(volume provisioning): fix CStor volume provisioning request without creating PVC [#1738](https://github.com/openebs/maya/pull/1738) ([mittachaitu](https://github.com/mittachaitu))

v2.1.0-RC2 / 2020-09-11
========================

 * No changes

v2.1.0-RC1 / 2020-09-08
========================

 * chore(upgrade): add valid current versions for 2.1.0 upgrades [#1749](https://github.com/openebs/maya/pull/1749) ([shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(spc): add capability to specify allowed BD tags on SPC [#1748](https://github.com/openebs/maya/pull/1748) ([sonasingh46](https://github.com/sonasingh46))
 * fix(spc): provision cStor stripe based pools with single raid group [#1744](https://github.com/openebs/maya/pull/1744) ([mittachaitu](https://github.com/mittachaitu))
 * fix(volume provisioning): fix CStor volume provisioning request without creating PVC [#1738](https://github.com/openebs/maya/pull/1738) ([mittachaitu](https://github.com/mittachaitu))

v1.12.1-RC1 / 2020-08-15
========================

 * refact(localpv): add ENV to allow skipping leader election [#1745](https://github.com/openebs/maya/pull/1745) ([prateekpandey14](https://github.com/prateekpandey14))
 * cherry-pick(fix): cherry-pick of PR #1738 [#1740](https://github.com/openebs/maya/pull/1740) ([mittachaitu](https://github.com/mittachaitu))
 * fix(version): fix castemplate output runstask names [#1736](https://github.com/openebs/maya/pull/1736) ([shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(version): use tagged version instead of VERSION file during build [#1734](https://github.com/openebs/maya/pull/1734) ([shubham14bajpai](https://github.com/shubham14bajpai))

v2.0.0 / 2020-08-13
========================

 * refact(localpv): add ENV to allow skipping leader election ([#1743](https://github.com/openebs/maya/pull/1743), [prateekpandey14](https://github.com/prateekpandey14))

v2.0.0-RC1 / 2020-08-11
======================== 

 * fix(volume provisioning): fix CStor volume provisioning request without creating PVC ([#1738](https://github.com/openebs/maya/pull/1738), [mittachaitu](https://github.com/mittachaitu))

v2.0.0-RC1 / 2020-08-08
========================

 * fix(build): use buildscript file `push` from openebs/charts ([#1737](https://github.com/openebs/maya/pull/1737), [shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(version): fix castemplate output runstask version suffix ([#1735](https://github.com/openebs/maya/pull/1735), [shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(version): use tagged version instead of VERSION file while build ([#1733](https://github.com/openebs/maya/pull/1733), [shubham14bajpai](https://github.com/shubham14bajpai))

v1.12.0 / 2020-07-14
========================

 * fix(upgrade): check for HA apisever pods instead of a single pod ([#1730](https://github.com/openebs/maya/pull/1730) ,[shubham14bajpai](https://github.com/shubham14bajpai))
 * refact(webhook): make webhook config failure policy configurable ([#1726](https://github.com/openebs/maya/pull/1726) ,[prateekpandey14](https://github.com/prateekpandey14))
 * chore(upgrade): add support for upgrading to any custom tag within same version ([#1724](https://github.com/openebs/maya/pull/1724) ,[shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(build): travis CI failure in forked Repo git error ([#1722](https://github.com/openebs/maya/pull/1722) ,[prateekpandey14](https://github.com/prateekpandey14))
 * fix(usage): add nil checks to avoid panic ([#1720](https://github.com/openebs/maya/pull/1720) ,[kmova](https://github.com/kmova))
 * fix(upgrade): increase wait for deployment rollout status ([#1719](https://github.com/openebs/maya/pull/1719) ,[shubham14bajpai](https://github.com/shubham14bajpai))

v1.12.0-RC2 / 2020-07-11
========================

 * No changes

v1.12.0-RC1 / 2020-07-08
========================

 * fix(upgrade): check for HA apisever pods instead of a single pod ([#1730](https://github.com/openebs/maya/pull/1730) ,[shubham14bajpai](https://github.com/shubham14bajpai))
 * refact(webhook): make webhook config failure policy configurable ([#1726](https://github.com/openebs/maya/pull/1726) ,[prateekpandey14](https://github.com/prateekpandey14))
 * chore(upgrade): add support for upgrading to any custom tag within same version ([#1724](https://github.com/openebs/maya/pull/1724) ,[shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(build): travis CI failure in forked Repo git error ([#1722](https://github.com/openebs/maya/pull/1722) ,[prateekpandey14](https://github.com/prateekpandey14))
 * fix(usage): add nil checks to avoid panic ([#1720](https://github.com/openebs/maya/pull/1720) ,[kmova](https://github.com/kmova))
 * fix(upgrade): increase wait for deployment rollout status ([#1719](https://github.com/openebs/maya/pull/1719) ,[shubham14bajpai](https://github.com/shubham14bajpai))

v1.11.0 / 2020-06-12
========================

 * fix(BDD): update checks after de-provisioning the CStor Volume ([#1715](https://github.com/openebs/maya/pull/1715), [mittachaitu](https://github.com/mittachaitu))
 * fix(build): remove duplicate declaration of constant ([#1714](https://github.com/openebs/maya/pull/1714), [prateekpandey14](https://github.com/prateekpandey14))
 * chore(Makefile): remove cspc,cvc,cspi based controller image builds ([#1713](https://github.com/openebs/maya/pull/1713), [prateekpandey14](https://github.com/prateekpandey14))
 * skip validations checks and handle snapshot deletion once migrated to CSI ([#1712](https://github.com/openebs/maya/pull/1712), [prateekpandey14](https://github.com/prateekpandey14))
 * chore(go-module): migrate vendor to go module ([#1711](https://github.com/openebs/maya/pull/1711), [vaniisgh](https://github.com/vaniisgh))
 * fix(spc): validate SPC deletion request ([#1710](https://github.com/openebs/maya/pull/1710), [mittachaitu](https://github.com/mittachaitu))
 * feat(usage): include pvc name in volume events ([#1708](https://github.com/openebs/maya/pull/1708), [kmova](https://github.com/kmova))
 * refactor(build): fix hard coding of image org (#1703) ([#1705](https://github.com/openebs/maya/pull/1705), [kmova](https://github.com/kmova))
 * feat(build): Automating localpv provisioner build for ppc64le ([#1704](https://github.com/openebs/maya/pull/1704), [Pensu](https://github.com/Pensu))
 * refactor(build): fix hard coding of image org ([#1703](https://github.com/openebs/maya/pull/1703), [kmova](https://github.com/kmova))
 * refact(exporter): handle concurrent scrape requests ([#1698](https://github.com/openebs/maya/pull/1698), [utkarshmani1997](https://github.com/utkarshmani1997))

v1.11.0-RC2 / 2020-06-11
========================

 * fix(BDD): update checks after de-provisioning the CStor Volume ([#1715](https://github.com/openebs/maya/pull/1715), [mittachaitu](https://github.com/mittachaitu))

v1.11.0-RC1 / 2020-06-09
========================

 * fix(build): remove duplicate declaration of constant ([#1714](https://github.com/openebs/maya/pull/1714), [prateekpandey14](https://github.com/prateekpandey14))
 * chore(Makefile): remove cspc,cvc,cspi based controller image builds ([#1713](https://github.com/openebs/maya/pull/1713), [prateekpandey14](https://github.com/prateekpandey14))
 * skip validations checks and handle snapshot deletion once migrated to CSI ([#1712](https://github.com/openebs/maya/pull/1712), [prateekpandey14](https://github.com/prateekpandey14))
 * chore(go-module): migrate vendor to go module ([#1711](https://github.com/openebs/maya/pull/1711), [vaniisgh](https://github.com/vaniisgh))
 * fix(spc): validate SPC deletion request ([#1710](https://github.com/openebs/maya/pull/1710), [mittachaitu](https://github.com/mittachaitu))
 * feat(usage): include pvc name in volume events ([#1708](https://github.com/openebs/maya/pull/1708), [kmova](https://github.com/kmova))
 * refactor(build): fix hard coding of image org (#1703) ([#1705](https://github.com/openebs/maya/pull/1705), [kmova](https://github.com/kmova))
 * feat(build): Automating localpv provisioner build for ppc64le ([#1704](https://github.com/openebs/maya/pull/1704), [Pensu](https://github.com/Pensu))
 * refactor(build): fix hard coding of image org ([#1703](https://github.com/openebs/maya/pull/1703), [kmova](https://github.com/kmova))
 * refact(exporter): handle concurrent scrape requests ([#1698](https://github.com/openebs/maya/pull/1698), [utkarshmani1997](https://github.com/utkarshmani1997))

v1.10.0 / 2020-05-15
========================

 * fix(webhook): cleanup old resources from 1.0.0 release ([#1696](https://github.com/openebs/maya/pull/1696),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(install): enable/disable crd installation([#1693](https://github.com/openebs/maya/pull/1693),
 [@kmova](https://github.com/kmova))
 * refact(build): make the docker images configurable ([#1680](https://github.com/openebs/maya/pull/1680),
 [@kmova](https://github.com/kmova))
 * fix(upgrade): fix version comparison function ([#1681](https://github.com/openebs/maya/pull/1681),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(maya-exporter): fix duplicate metrics for openebs_zpool_last_sync_time ([#1678](https://github.com/openebs/maya/pull/1678)),
 [@slalwani97](https://github.com/slalwani97))
 * fix(cstor-restore): fixing restore api to return failure if cstorrestore is in invalid state ([#1682](https://github.com/openebs/maya/pull/1682),
 [@mynktl](https://github.com/mynktl))
 * refact(webhook): update webhookconfiguration failure policy to Fail ([#1672](https://github.com/openebs/maya/pull/1672),
 [@prateekpandey14 ](https://github.com/prateekpandey14))
 * fix(cStorvolumereplica): record the error while fetching the cvr status ([#1675](https://github.com/openebs/maya/pull/1675),
 [@mittachaitu](https://github.com/mittachaitu))
 * chore(doc): add unreleased changelogs in repo ([#1674](https://github.com/openebs/maya/pull/1674),
 [@mittachaitu](https://github.com/mittachaitu))
 * docs(contributor): update project contribution guidelines ([#1673](https://github.com/openebs/maya/pull/1673),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(CVR): enable REBUILD_ESTIMATES feature gate ([#1670](https://github.com/openebs/maya/pull/1670),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(webhook): reject PVC deletion request when dependent snapshots exists ([#1669](https://github.com/openebs/maya/pull/1669),
 [@mittachaitu](https://github.com/mittachaitu))

v1.10.0-RC2 / 2020-05-13
========================

 * feat(install): enable/disable crd installation([#1693](https://github.com/openebs/maya/pull/1693),
 [@kmova](https://github.com/kmova))
 * refact(build): make the docker images configurable ([#1680](https://github.com/openebs/maya/pull/1680),
 [@kmova](https://github.com/kmova))

v1.10.0-RC1 / 2020-05-08
========================

 * fix(upgrade): fix version comparison function ([#1681](https://github.com/openebs/maya/pull/1681),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(maya-exporter): fix duplicate metrics for openebs_zpool_last_sync_time ([#1678](https://github.com/openebs/maya/pull/1678)),
 [@slalwani97](https://github.com/slalwani97))
 * fix(cstor-restore): fixing restore api to return failure if cstorrestore is in invalid state ([#1682](https://github.com/openebs/maya/pull/1682),
 [@mynktl](https://github.com/mynktl))
 * refact(webhook): update webhookconfiguration failure policy to Fail ([#1672](https://github.com/openebs/maya/pull/1672),
 [@prateekpandey14 ](https://github.com/prateekpandey14))
 * fix(cStorvolumereplica): record the error while fetching the cvr status ([#1675](https://github.com/openebs/maya/pull/1675),
 [@mittachaitu](https://github.com/mittachaitu))
 * chore(doc): add unreleased changelogs in repo ([#1674](https://github.com/openebs/maya/pull/1674),
 [@mittachaitu](https://github.com/mittachaitu))
 * docs(contributor): update project contribution guidelines ([#1673](https://github.com/openebs/maya/pull/1673),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(CVR): enable REBUILD_ESTIMATES feature gate ([#1670](https://github.com/openebs/maya/pull/1670),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(webhook): reject PVC deletion request when dependent snapshots exists ([#1669](https://github.com/openebs/maya/pull/1669),
 [@mittachaitu](https://github.com/mittachaitu))

v1.9.0 / 2020-04-14
========================

 * fix(jiva): add tolerations to jiva cleanup jobs([#1667](https://github.com/openebs/maya/pull/1667),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(build): update alpine base images to 3.11.5 ([#1660](https://github.com/openebs/maya/pull/1660),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(localpv-hostpath): pv deletion caused panic when node not found ([#1662](https://github.com/openebs/maya/pull/1662),
 [@kmova](https://github.com/kmova))
 * feat(build): enable automatic building and pushing of arm images from ci ([#1650](https://github.com/openebs/maya/pull/1650),
 [@akhilerm](https://github.com/akhilerm))
 * chore(build): trigger downstream repo release ([#1657](https://github.com/openebs/maya/pull/1657),
 [@kmova](https://github.com/kmova))
 * feat(upgrade): enable bulk upgrade for volumes and spc ([#1655](https://github.com/openebs/maya/pull/1655),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(localpv-device): allow local pv device on select devices ([#1648](https://github.com/openebs/maya/pull/1648),
 [@kmova](https://github.com/kmova))
 * feat(build): adding pc64le build for local provisioner ([#1632](https://github.com/openebs/maya/pull/1632),
 [@Pensu](https://github.com/Pensu))
 * feat(upgrade): split jiva replicas and migrate jiva resources to openebs ([#1646](https://github.com/openebs/maya/pull/1646),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(cstorBackup, delete): support for snapshot deletion, created by velero-plugin ([#1644](https://github.com/openebs/maya/pull/1644),
 [@mynktl](https://github.com/mynktl))
 * feat(BDD): add positive test cases for verifying waitforfirstconsumer With CStor Volume Provisioning ([#1643](https://github.com/openebs/maya/pull/1643),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(estimate_rebuilds): add pending snapshots in CVR Status by talking to peer replicas ([#1641](https://github.com/openebs/maya/pull/1641),
 [@mittachaitu](https://github.com/mittachaitu))
 * refact(jiva): create separate replica deployments and move all resources to openebsNamespace ([#1636](https://github.com/openebs/maya/pull/1636),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(snapList): add snapshots information in CVR Status ([#1639](https://github.com/openebs/maya/pull/1639),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(local-snapshot-restore, velero) : support to restore CStor snapshot to different namespace using velero ([#1575](https://github.com/openebs/maya/pull/1575),
 [@mynktl](https://github.com/mynktl))
 * refact(log): remove cstor prefix from localPV log and alert messages ([#1638](https://github.com/openebs/maya/pull/1638),
 [@akhilerm](https://github.com/akhilerm))
 * fix(provisioning): support to provision volumes in case of WaitForFirstConsumer ([#1637](https://github.com/openebs/maya/pull/1637),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(csp): add pool protection finalizer on CSP ([#1635](https://github.com/openebs/maya/pull/1635),
 [@mittachaitu](https://github.com/mittachaitu))

v1.9.0-RC2 / 2020-04-12
========================

 * fix(jiva): add tolerations to jiva cleanup jobs([#1667](https://github.com/openebs/maya/pull/1667),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(build): update alpine base images to 3.11.5 ([#1660](https://github.com/openebs/maya/pull/1660),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(localpv-hostpath): pv deletion caused panic when node not found ([#1662](https://github.com/openebs/maya/pull/1662),
 [@kmova](https://github.com/kmova))
 * feat(build): enable automatic building and pushing of arm images from ci ([#1650](https://github.com/openebs/maya/pull/1650),
 [@akhilerm](https://github.com/akhilerm))


v1.9.0-RC1 / 2020-04-07
========================

 * chore(build): trigger downstream repo release ([#1657](https://github.com/openebs/maya/pull/1657),
 [@kmova](https://github.com/kmova))
 * feat(upgrade): enable bulk upgrade for volumes and spc ([#1655](https://github.com/openebs/maya/pull/1655),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(localpv-device): allow local pv device on select devices ([#1648](https://github.com/openebs/maya/pull/1648),
 [@kmova](https://github.com/kmova))
 * feat(build): adding pc64le build for local provisioner ([#1632](https://github.com/openebs/maya/pull/1632),
 [@Pensu](https://github.com/Pensu))
 * feat(upgrade): split jiva replicas and migrate jiva resources to openebs ([#1646](https://github.com/openebs/maya/pull/1646),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(cstorBackup, delete): support for snapshot deletion, created by velero-plugin ([#1644](https://github.com/openebs/maya/pull/1644),
 [@mynktl](https://github.com/mynktl))
 * feat(BDD): add positive test cases for verifying waitforfirstconsumer With CStor Volume Provisioning ([#1643](https://github.com/openebs/maya/pull/1643),
 [@mittachaitu](https://github.com/mittachaitu))
 * chore(BDD): checks to verify pool protectionfinalizer ([#1640](https://github.com/openebs/maya/pull/1640),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(estimate_rebuilds): add pending snapshots on CVR by talking to peer replicas ([#1641](https://github.com/openebs/maya/pull/1641),
 [@mittachaitu](https://github.com/mittachaitu))
 * refact(jiva): create separate replica deployments and move all resources to openebsNamespace ([#1636](https://github.com/openebs/maya/pull/1636),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(snapList): add snapshots information on CVR ([#1639](https://github.com/openebs/maya/pull/1639),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(local-snapshot-restore, velero) : support to restore CStor snapshot to different namespace using velero ([#1575](https://github.com/openebs/maya/pull/1575),
 [@mynktl](https://github.com/mynktl))
 * refact(log): remove cstor prefix from localPV log and alert messages ([#1638](https://github.com/openebs/maya/pull/1638),
 [@akhilerm](https://github.com/akhilerm))
 * fix(provisioning): support to provision volumes in case of WaitForFirstConsumer ([#1637](https://github.com/openebs/maya/pull/1637),
 [@mittachaitu](https://github.com/mittachaitu))


v1.8.0 / 2020-03-14
========================

 * fix(upgrade): increase timeout for httpClient([#1630](https://github.com/openebs/maya/pull/1630),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * refact(upgrade): scale down jiva target deploy before replica patch ([#1626](https://github.com/openebs/maya/pull/1626),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(validation): add webhook validations for CVC replica scale  ([#1621](https://github.com/openebs/maya/pull/1621),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(backup, cstor): fetching correct snap name in case of base backup failure ([#1622](https://github.com/openebs/maya/pull/1622),
 [@mynktl](https://github.com/mynktl))
 * refact(upgrade): add support for 1.8.0 upgrades ([#1624](https://github.com/openebs/maya/pull/1624),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(version): bump master to 1.8.0 and Go version to 1.12.16 ([#1614](https://github.com/openebs/maya/pull/1614),
 [@prateekpandey14](https://github.com/prateekpandey14))
 * feat(csp, poolROThreshold): setting default poolROThreshold to 85 for CSP ([#1623](https://github.com/openebs/maya/pull/1623),
 [@mynktl](https://github.com/mynktl))
 * feat(spc,csp): adding support for pool ReadOnly Threshold limit ([#1609](https://github.com/openebs/maya/pull/1609),
 [@mynktl](https://github.com/mynktl))
 * feat(cvc-operator): add automatic scaling of volumereplicas for CSI volumes ([#1613](https://github.com/openebs/maya/pull/1613),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(exporter): handle pool sync time metrics collection gracefully ([#1616](https://github.com/openebs/maya/pull/1616),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(travis): add check to verify existence of udev ([#1619](https://github.com/openebs/maya/pull/1619),
 [@mittachaitu](https://github.com/mittachaitu))

v1.8.0-RC2 / 2020-03-14
========================

 * refact(upgrade): scale down jiva target deploy before replica patch ([#1626](https://github.com/openebs/maya/pull/1626),
 [@shubham14bajpai](https://github.com/shubham14bajpai))

v1.8.0-RC1 / 2020-03-06
========================

 * feat(validation): add webhook validations for CVC replica scale  ([#1621](https://github.com/openebs/maya/pull/1621),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(backup, cstor): fetching correct snap name in case of base backup failure ([#1622](https://github.com/openebs/maya/pull/1622),
 [@mynktl](https://github.com/mynktl))
 * refact(upgrade): add support for 1.8.0 upgrades ([#1624](https://github.com/openebs/maya/pull/1624),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(version): bump master to 1.8.0 and Go version to 1.12.16 ([#1614](https://github.com/openebs/maya/pull/1614),
 [@prateekpandey14](https://github.com/prateekpandey14))
 * feat(csp, poolROThreshold): setting default poolROThreshold to 85 for CSP ([#1623](https://github.com/openebs/maya/pull/1623),
 [@mynktl](https://github.com/mynktl))
 * feat(spc,csp): adding support for pool ReadOnly Threshold limit ([#1609](https://github.com/openebs/maya/pull/1609),
 [@mynktl](https://github.com/mynktl))
 * feat(cvc-operator): add automatic scaling of volumereplicas for CSI volumes ([#1613](https://github.com/openebs/maya/pull/1613),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(exporter): handle pool sync time metrics collection gracefully ([#1616](https://github.com/openebs/maya/pull/1616),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(travis): add check to verify existence of udev ([#1619](https://github.com/openebs/maya/pull/1619),
 [@mittachaitu](https://github.com/mittachaitu))

v1.7.0-RC1 / 2020-02-07
========================

 * fix(BDD): wait for restarted pod to come to running state ([#1608](https://github.com/openebs/maya/pull/1608),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(BDD): fix BDD test case by increasing retries ([#1607](https://github.com/openebs/maya/pull/1607),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(jiva-cleanup): allow cleanup jobs to run as previleged mode ([#1600](https://github.com/openebs/maya/pull/1600),
 [@utkarshmani1997](https://github.com/utkarshmani1997))
 * refact(upgrade): add volumes and mountpaths for cstor pools and volumes ([#1584](https://github.com/openebs/maya/pull/1584),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(new cStor deployments): add OpenEBS base directory in new CStor deployments ([#1599](https://github.com/openebs/maya/pull/1599),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(validation): using label instead of annotation in CSPC validation ([#1598](https://github.com/openebs/maya/pull/1598),
 [@mittachaitu](https://github.com/mittachaitu))
 * chore(bump-version): bump kubernetes dependency to "kubernetes-1.17.0" version ([#1596](https://github.com/openebs/maya/pull/1596),
 [@shovanmaity](https://github.com/shovanmaity))
 * feat(cvc,csi): pre provisioning and prioritize pool list for replica scheduling ([#1591](https://github.com/openebs/maya/pull/1591),
 [@prateekpandey14](https://github.com/prateekpandey14))
 * chore(spc):introduce new field to control overprovisioning ([#1597](https://github.com/openebs/maya/pull/1597),
 ([#1602](https://github.com/openebs/maya/pull/1602),
 [@chandankumar4](https://github.com/chandankumar4))
 * feat(PDB, cStor Pools): add a support to create PDB for cStor ([#1573](https://github.com/openebs/maya/pull/1573),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(cspc, webhook): add validation for cspc deletion ([#1594](https://github.com/openebs/maya/pull/1594),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(install): remove old castemplates and runtask after upgrade ([#1595](https://github.com/openebs/maya/pull/1595),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * refactor(gitlab-yaml): removed the request to trigger pipeline in packet ([#1592](https://github.com/openebs/maya/pull/1592),
 [@nsathyaseelan](https://github.com/nsathyaseelan))
 * fix(pool-mgmt): removing stale CVR/CSP when dataset/pool does not exist ([#1587](https://github.com/openebs/maya/pull/1587),
 [@mynktl](https://github.com/mynktl))
 * fix(cspc): cleanup bdcs after deletion of CSPI or removing poolSpec from CSPC ([#1579](https://github.com/openebs/maya/pull/1579),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(core): handle mount paths ([#1589](https://github.com/openebs/maya/pull/1589),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(spc): add feature to limit overprovisioning of cstor volumes ([#1577](https://github.com/openebs/maya/pull/1577),
 ([#1593](https://github.com/openebs/maya/pull/1593),
 [@sonasingh46](https://github.com/sonasingh46))
 * fix(cStor, deployments): add OpenEBS base directory in deployments ([#1583](https://github.com/openebs/maya/pull/1583),
 ([#1605](https://github.com/openebs/maya/pull/1605),
 ([#1603](https://github.com/openebs/maya/pull/1603),
 ([#1601](https://github.com/openebs/maya/pull/1601),
 [@mittachaitu](https://github.com/mittachaitu))

