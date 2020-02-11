# Recovery
The data in the tenant zookeepers is persisted on AWS volumes.
In case these volumes are no longer available, the data can be recovered from the burry backup, stored on S3. Here are the
instructions to do so.
## Stop services
* make sure all related services are stopped:
  * dsh/burry
  * dsh/nicodemus/prov-faas
  * dsh/nicodemus/prov-app
  * any tenant FAAS services (look for `flink-cluster-jobmanager` and `flink-cluster-taskmanager`).

    _Note:_ Make sure prov-app has been stopped before trying to stop the FAAS services.
## Prepare the backup file
* Download the correct version of the backup from the S3 bucket.
  * Location: `<bucket>/environments/<platform>/tenantzookeeper/backup.burry.zip`

    e.g.: `environments-dsh-eu-central-1/environments/dev-dsh/tenantzookeeper/backup.burry.zip`
  * Version: the bucket is versioned, so make sure when downloading you select the correct version (timestamp) from where you need to recover
* Rename the backup: `burry` expects that the name of the backup file corresponds to the sessionid it has stored in the backup. The easiest way to get the sessionid is:
  * execute the cmd: `unzip -l backup.burry.zip | head -4`
    ```
    Archive:  backup.burry.zip
    Length      Date    Time    Name
    ---------  ---------- -----   ----
          233  00-00-1980 00:00   1579852653/.burrymeta
    ```
  * the sessionid is the number before `./burrymeta`. (i.e. `1579852653`)
  * rename the file: `mv backup.burry.zip <sessionid>.zip`
## Build burry
* build the `burry` executable. See the instructions in [README.md](README.md)
## Recover zookeeper data
* scp both the executable, and the zip file to a dsh node
* ssh into the node
* run the command:
  `./burry.sh -o restore -e tenant-zookeeper-1.dsh.marathon.mesos:2181 -t local -s <sessionid>`
## Restart services
You can now safely resume the `dsh` services:
* dsh/burry
* dsh/nicodemus/prov-faas
* dsh/nicodemus/prov-app

_NOTE:_ It is not necessary to restart all the tenant flink services. They should be started automatically by the Nicodemus application provisioner.