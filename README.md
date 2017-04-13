# DashVid-Uploader
A utility to upload video files to DashVid.io

# Concourse

To load the build into Concourse run the following commands

`fly -t jassoft login -n JassoftLtd -c CONCOURSE_URL`

`fly -t jassoft set-pipeline -p DashVid-Uploader -c DashVid-Uploader.yml --load-vars-from secrets.yml`

`fly -t jassoft unpause-pipeline --pipeline DashVid-Uploader`
