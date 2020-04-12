v1.9.0 / 2020-15-07
========================

 * fix(jiva): add tolerations to jiva cleanup jobs([#1667](https://github.com/openebs/maya/pull/1667),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(build): update alpine base images to 3.11.5 ([#1660](https://github.com/openebs/maya/pull/1660),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(upgarde): scale down jiva contoller deployment only if not upgraded ([#1663](https://github.com/openebs/maya/pull/1663),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(localpv-hostpath): pv deletion caused panic when node not found ([#1662](https://github.com/openebs/maya/pull/1662),
 [@kmova](https://github.com/kmova))
 * feat(build): enable arm64 auto build ([#1650](https://github.com/openebs/maya/pull/1650),
 [@akhilerm](https://github.com/akhilerm))
 * chore(build): trigger downstream repo release ([#1657](https://github.com/openebs/maya/pull/1657),
 [@kmova](https://github.com/kmova))
 * feat(upgrade): enable bulk upgrade for volumes and spc ([#1655](https://github.com/openebs/maya/pull/1655),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(localpv-device): allow local pv device on select devices ([#1648](https://github.com/openebs/maya/pull/1648),
 [@kmova](https://github.com/kmova))
 * feat(build): adding pc64le build for local provisioner ([#1632](https://github.com/openebs/maya/pull/1632),
 [@Pensu](https://github.com/Pensu))
 * fix(BDD): fix SPC reconciliation BDD by adding extra filter ([#1654](https://github.com/openebs/maya/pull/1654),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(upgrade): split jiva replicas and migrate jiva resources to openebs ([#1646](https://github.com/openebs/maya/pull/1646),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(cstorBackup, delete): support for snapshot deletion ([#1644](https://github.com/openebs/maya/pull/1644),
 [@mynktl](https://github.com/mynktl))
 * fix(bdd): fix BDD to avoid failures during creation of StorageClass ([#1652](https://github.com/openebs/maya/pull/1652),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(jiva): add namespace for stateful set target affinity([#1651](https://github.com/openebs/maya/pull/1651),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(bdd): update the namespace for verifying jiva pods ([#1649](https://github.com/openebs/maya/pull/1649),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(api): add label selector in BDC spec ([#1647](https://github.com/openebs/maya/pull/1647),
 [@akhilerm](https://github.com/akhilerm))
 * feat(BDD): add positive test cases for verifying waitforfirstconsumer With CStor Volume Provisioning ([#1643](https://github.com/openebs/maya/pull/1643),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(jiva): add namespace to podAffinity for target([#1645](https://github.com/openebs/maya/pull/1645),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(BDD): checks to verify pool protectionfinalizer ([#1640](https://github.com/openebs/maya/pull/1640),
 [@mittachaitu](https://github.com/mittachaitu))
 *  feat(estimate_rebuilds): add pending snapshots on CVR by talking to peer replicas ([#1641](https://github.com/openebs/maya/pull/1641),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(apis): add new fields into BlockDevice([#1642](https://github.com/openebs/maya/pull/1642),
 [@akhilerm](https://github.com/akhilerm))
 * refact(jiva): create separate replica deployments and move all resources to openebsNamespace ([#1636](https://github.com/openebs/maya/pull/1636),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(snapList): add snapshots information on CVR ([#1639](https://github.com/openebs/maya/pull/1639),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(local-snapshot-restore, velero) : support to restore local snapshot to different namespace using velero ([#1575](https://github.com/openebs/maya/pull/1575),
 [@mynktl](https://github.com/mynktl))
 * refact(log): remove cstor prefix from localPV log and alert messages ([#1638](https://github.com/openebs/maya/pull/1638),
 [@akhilerm](https://github.com/akhilerm))
 * fix(provisioning): support to provision volumes in case of WaitForFirstConsumer ([#1637](https://github.com/openebs/maya/pull/1637),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(csp): add pool protection finalizer on CSP ([#1635](https://github.com/openebs/maya/pull/1635),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(restore,pvc): handling of pvc's annotation for existing velero-plugin's restore ([#1631](https://github.com/openebs/maya/pull/1631),
 [@mynktl](https://github.com/mynktl))
 * fix(apiserver,volume) Removing PVC dependency from volume creation path ([#1570](https://github.com/openebs/maya/pull/1570),
 [@mynktl](https://github.com/mynktl))

v1.9.0-RC2 / 2020-12-07
========================

 * fix(jiva): add tolerations to jiva cleanup jobs([#1667](https://github.com/openebs/maya/pull/1667),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(build): update alpine base images to 3.11.5 ([#1660](https://github.com/openebs/maya/pull/1660),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(upgarde): scale down jiva contoller deployment only if not upgraded ([#1663](https://github.com/openebs/maya/pull/1663),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(localpv-hostpath): pv deletion caused panic when node not found ([#1662](https://github.com/openebs/maya/pull/1662),
 [@kmova](https://github.com/kmova))
 * feat(build): enable arm64 auto build ([#1650](https://github.com/openebs/maya/pull/1650),
 [@akhilerm](https://github.com/akhilerm))


v1.9.0-RC1 / 2020-05-07
========================

 * chore(build): trigger downstream repo release ([#1657](https://github.com/openebs/maya/pull/1657),
 [@kmova](https://github.com/kmova))
 * feat(upgrade): enable bulk upgrade for volumes and spc ([#1655](https://github.com/openebs/maya/pull/1655),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(localpv-device): allow local pv device on select devices ([#1648](https://github.com/openebs/maya/pull/1648),
 [@kmova](https://github.com/kmova))
 * feat(build): adding pc64le build for local provisioner ([#1632](https://github.com/openebs/maya/pull/1632),
 [@Pensu](https://github.com/Pensu))
 * fix(BDD): fix SPC reconciliation BDD by adding extra filter ([#1654](https://github.com/openebs/maya/pull/1654),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(upgrade): split jiva replicas and migrate jiva resources to openebs ([#1646](https://github.com/openebs/maya/pull/1646),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(cstorBackup, delete): support for snapshot deletion ([#1644](https://github.com/openebs/maya/pull/1644),
 [@mynktl](https://github.com/mynktl))
 * fix(bdd): fix BDD to avoid failures during creation of StorageClass ([#1652](https://github.com/openebs/maya/pull/1652),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(jiva): add namespace for stateful set target affinity([#1651](https://github.com/openebs/maya/pull/1651),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(bdd): update the namespace for verifying jiva pods ([#1649](https://github.com/openebs/maya/pull/1649),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(api): add label selector in BDC spec ([#1647](https://github.com/openebs/maya/pull/1647),
 [@akhilerm](https://github.com/akhilerm))
 * feat(BDD): add positive test cases for verifying waitforfirstconsumer With CStor Volume Provisioning ([#1643](https://github.com/openebs/maya/pull/1643),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(jiva): add namespace to podAffinity for target([#1645](https://github.com/openebs/maya/pull/1645),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * chore(BDD): checks to verify pool protectionfinalizer ([#1640](https://github.com/openebs/maya/pull/1640),
 [@mittachaitu](https://github.com/mittachaitu))
 *  feat(estimate_rebuilds): add pending snapshots on CVR by talking to peer replicas ([#1641](https://github.com/openebs/maya/pull/1641),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(apis): add new fields into BlockDevice([#1642](https://github.com/openebs/maya/pull/1642),
 [@akhilerm](https://github.com/akhilerm))
 * refact(jiva): create separate replica deployments and move all resources to openebsNamespace ([#1636](https://github.com/openebs/maya/pull/1636),
 [@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(snapList): add snapshots information on CVR ([#1639](https://github.com/openebs/maya/pull/1639),
 [@mittachaitu](https://github.com/mittachaitu))
 * feat(local-snapshot-restore, velero) : support to restore local snapshot to different namespace using velero ([#1575](https://github.com/openebs/maya/pull/1575),
 [@mynktl](https://github.com/mynktl))
 * refact(log): remove cstor prefix from localPV log and alert messages ([#1638](https://github.com/openebs/maya/pull/1638),
 [@akhilerm](https://github.com/akhilerm))
 * fix(provisioning): support to provision volumes in case of WaitForFirstConsumer ([#1637](https://github.com/openebs/maya/pull/1637),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(csp): add pool protection finalizer on CSP ([#1635](https://github.com/openebs/maya/pull/1635),
 [@mittachaitu](https://github.com/mittachaitu))
 * fix(restore,pvc): handling of pvc's annotation for existing velero-plugin's restore ([#1631](https://github.com/openebs/maya/pull/1631),
 [@mynktl](https://github.com/mynktl))
 * fix(apiserver,volume) Removing PVC dependency from volume creation path ([#1570](https://github.com/openebs/maya/pull/1570),
 [@mynktl](https://github.com/mynktl))


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
  [#1602](https://github.com/openebs/maya/pull/1602),
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
  [#1593](https://github.com/openebs/maya/pull/1593),
  [@sonasingh46](https://github.com/sonasingh46))
  * fix(cStor, deployments): add OpenEBS base directory in deployments ([#1583](https://github.com/openebs/maya/pull/1583),
  [#1605](https://github.com/openebs/maya/pull/1605),
  [#1603](https://github.com/openebs/maya/pull/1603),
  [#1601](https://github.com/openebs/maya/pull/1601),
  [@mittachaitu](https://github.com/mittachaitu))
