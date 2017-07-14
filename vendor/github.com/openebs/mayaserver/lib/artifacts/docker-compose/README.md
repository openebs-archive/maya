## docker-compose file

```bash
# Create the maya api server container
sudo docker-compose up -d

# Check the maya api server container logs
sudo docker-compose logs

# Alternative
sudo docker logs <Container-ID>

# verify the m-apiserver service within this container
curl http://<Container-IP>:5656/latest/meta-data/instance-id

# Verify INI file inside the container
amit:docker$ sudo docker exec -it <Container-ID> bash
root@921f974ee490:/# 
root@921f974ee490:/# cat /etc/mayaserver/orchprovider/nomad_global.INI
```

### Internals of the compose file

- Use the right maya api server docker image

```yaml
mapiserver:
  image: openebs/m-apiserver:0.2-RC4
```

- Modify the ENV values based on the environment & requirements

```yaml
  environment:
    - ENV MAYA_API_SERVER_VERSION="0.2-RC4"
    - NOMAD_ADDR="http:\/\/172.28.128.3:4646"
    - NOMAD_CN_TYPE="host"
    - NOMAD_CN_NETWORK_CIDR="172.28.128.1\/24"
    - NOMAD_CN_INTERFACE="enp0s8"
    - NOMAD_CS_PERSISTENCE_LOCATION="\/tmp\/"
    - NOMAD_CS_REPLICA_COUNT="2"
```

## Cleaning up Docker Images & Containers

> This might be required in your dev/test machine or your own laptop.

```bash
# Stop all Containers
sudo docker stop $(sudo docker ps -a -q)

# Remove all Containers & its Associated Volumes
sudo docker rm -v $(sudo docker ps -a -q)

# Remove all Images
sudo docker rmi $(sudo docker images -q)
```
