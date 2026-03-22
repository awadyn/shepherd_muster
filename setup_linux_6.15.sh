git clone git@github.com:awadyn/ixgbe.git
wget https://www.kernel.org/pub/linux/kernel/v6.x/linux-6.15.1.tar.gz
tar -xzf linux-6.15.1.tar.gz
cp -r ixgbe/ linux-6.15.1/drivers/net/ethernet/intel/

echo 
echo

sudo apt install -y fakeroot dwarves flex bison libssl-dev libelf-dev

echo
echo

cd linux-6.15.1/
cp -v /boot/config-$(uname -r) .config 
yes "" | make localmodconfig
scripts/config --disable SYSTEM_TRUSTED_KEYS
scripts/config --disable SYSTEM_REVOCATION_KEYS
scripts/config --set-str CONFIG_SYSTEM_TRUSTED_KEYS ""
scripts/config --set-str CONFIG_SYSTEM_REVOCATION_KEYS ""
scripts/config --enable CONFIG_X86_MSR

echo
echo

fakeroot make -j10
sudo make modules_install
sudo make install
sudo reboot

