#!/bin/sh
cd /tmp

/app/burry.sh -e ${ZOO_SERVERS:?} -i zk -t s3 -o daemon \
-p ${POLLTIME:?} \
${BLACKLIST:+-l "${BLACKLIST}"} \
-c s3.amazonaws.com,ACCESS_KEY_ID=${AWS_ACCESS_KEY:?},SECRET_ACCESS_KEY=${AWS_SECRET_KEY:?},BUCKET=${AWS_BUCKET:?},PREFIX=${AWS_PREFIX:?},SSL=true,OBJECT=${OBJECT:-backup.burry.zip}
