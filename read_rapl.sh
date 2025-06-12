sudo rdmsr -p 0 0x639 >> rapl_pkg.txt
sudo rdmsr -p 10 0x639 >> rapl_pkg.txt

sudo rdmsr -p 0 0x611 >> rapl_pp0.txt
sudo rdmsr -p 10 0x611 >> rapl_pp0.txt

