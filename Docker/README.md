```
docker run \
-v /path/to/postgres-ci/worker/config.yaml:/etc/postgres-ci/worker.yaml \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /tmp/postgres-ci/:/tmp/postgres_ci/ \
--name='postgres-ci-worker'  \
-d postgres-ci-worker
```


my local worker 

```
docker run \
-v /home/kshvakov/gosrc/src/github.com/postgres-ci/worker/config.sample.yaml:/etc/postgres-ci/worker.yaml \  <- worker konfig
-v /var/run/docker.sock:/var/run/docker.sock \
-v /home/kshvakov/testrepo:/home/kshvakov/testrepo \ <- if you use local repo
-v /tmp/postgres-ci/:/tmp/postgres-ci/ \ <- build dir 
--name='postgres-ci-worker' \
-d postgres-ci-worker
```
