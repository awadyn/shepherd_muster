node_list=( "$@" )
myself="awadyn"

for node in ${node_list[*]}; do
ssh $myself@$node 'sudo apt update && sudo apt upgrade -y;			\
	sudo apt install -y libevent-dev gengetopt libzmq3-dev;			\
	sudo apt install -y python2.7;						\
	wget -P ~/.local/lib https://bootstrap.pypa.io/pip/2.7/get-pip.py;	\
	python2.7 ~/.local/lib/get-pip.py --user;				\
	python2.7 -m pip install --user scons;					\
	git clone https://github.com/awadyn/mutilate.git;			\
	cd mutilate && python2.7 ~/.local/bin/scons;				\
	sudo rmmod mlx4_ib;							\
	sudo rmmod mlx4_core'
done

