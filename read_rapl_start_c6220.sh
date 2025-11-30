echo > rapl_pkg.txt
echo > rapl_pp0.txt

sudo rdmsr -p 0 0x611 >> rapl_pkg.txt
sudo rdmsr -p 8 0x611 >> rapl_pkg.txt

sudo rdmsr -p 0 0x639 >> rapl_pp0.txt
sudo rdmsr -p 8 0x639 >> rapl_pp0.txt

