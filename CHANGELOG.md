v1.7.0-RC1 / 2020-02-07
========================

  * fix(BDD): wait for restarted pod to come to running state ([#1608](https://github.com/openebs/maya/pull/1608),
  [@shubham14bajpai](https://github.com/shubham14bajpai))
  * fix(BDD): fix BDD test case ([#1607](https://github.com/openebs/maya/pull/1607),
  [@mittachaitu](https://github.com/mittachaitu))
  * fix(mountpath): fix sock file mount path in pool deployments ([#1605](https://github.com/openebs/maya/pull/1605),
  [@mittachaitu](https://github.com/mittachaitu))
  * fix(jiva-cleanup): allow cleanup jobs to run as previleged mode ([#1600](https://github.com/openebs/maya/pull/1600),
  [@utkarshmani1997](https://github.com/utkarshmani1997))
  * patch(upgrade): add volumes and mountpaths for cstor pools and volumes ([#1584](https://github.com/openebs/maya/pull/1584),
  [@shubham14bajpai](https://github.com/shubham14bajpai))
  * fix(sockfile): mount sockfile at empty direcotory ([#1603](https://github.com/openebs/maya/pull/1603),
  [@mittachaitu](https://github.com/mittachaitu))
  * test(spc): fix the test related to the PR #1597 ([#1602](https://github.com/openebs/maya/pull/1602),
  [@chandankumar4](https://github.com/chandankumar4))
  * fix(new cStor deployments): add OpenEBS base directory in new CStor deployments ([#1599](https://github.com/openebs/maya/pull/1599),
  [@mittachaitu](https://github.com/mittachaitu))
  * fix(sockfile): mount sockfile on emtyDir ([#1601](https://github.com/openebs/maya/pull/1601),
  [@mittachaitu](https://github.com/mittachaitu))
  * fix(validation): using label instead of annotation in CSPC validation ([#1598](https://github.com/openebs/maya/pull/1598),
  [@mittachaitu](https://github.com/mittachaitu))
  * chore(bump-version): bump kubernetes dependency to "kubernetes-1.17.0" version ([#1596](https://github.com/openebs/maya/pull/1596),
  [@shovanmaity](https://github.com/shovanmaity))
  * feat(cvc,csi): pre provisioning and prioritize pool list for replica scheduling ([#1591](https://github.com/openebs/maya/pull/1591),
  [@prateekpandey14](https://github.com/prateekpandey14))
  * chore(spc):introduce new field to control overprovisioning ([#1597](https://github.com/openebs/maya/pull/1597),
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
  * fix(overprovisioning): fix overprovisioning bug ([#1593](https://github.com/openebs/maya/pull/1593),
  [@sonasingh46](https://github.com/sonasingh46))
  * fix(core): handle mount paths ([#1589](https://github.com/openebs/maya/pull/1589),
  [@mittachaitu](https://github.com/mittachaitu))
  * feat(spc): add feature to limit overprovisioning of cstor volumes ([#1577](https://github.com/openebs/maya/pull/1577),
  [@sonasingh46](https://github.com/sonasingh46))
  * chore(version): bump version to 1.7.0 ([#1586](https://github.com/openebs/maya/pull/1586),
  [@shubham14bajpai](https://github.com/shubham14bajpai))
  * fix(cStor, deployments): add OpenEBS base directory in deployments ([#1583](https://github.com/openebs/maya/pull/1583),
  [@mittachaitu](https://github.com/mittachaitu))
