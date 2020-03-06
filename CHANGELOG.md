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

