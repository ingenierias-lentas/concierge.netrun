#mkdir mycontainer
#cd mycontainer
#docker create --name cont1 alpine sh
#docker export cont1 > alpine.tar
#docker rm -f cont1
#mkdir rootfs; tar -xf alpine.tar -C myrootfs
#runc spec
#sudo runc start # enter container
