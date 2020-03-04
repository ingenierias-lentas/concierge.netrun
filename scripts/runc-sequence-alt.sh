#mkdir mycontainer
#cd mycontainer
#mkdir rootfs
#docker export $(docker create busybox) | tar -C rootfs -xvf -
#runc spec
#sed -i 's;"sh";"top";' config.json # change start/launch parameter
#runc run container1 # run as container1
