# How to run
* Install java  
`$ sudo apt update && apt install openjdk-11-jdk`
* Install xvfb  
`$ sudo apt install xvfb`
* Unzip `welcomegram.zip`  
`$ unzip welcomegram.zip`
* Go to unpacked folder `build` and make file `config.json` from `config.json.dist`, change credentials and message which will be send to new subscribers  
`$ cd build && cp config.json.dist config.json`
* Run`welcomegram` file  
`$ ./welcomegram`
