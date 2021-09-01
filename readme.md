# armada-nmap

Parses the armada output and run the nmap service scan per IP for all open ports


### Installation
```
GO111MODULE=on go get github.com/k-sau/armada-nmap@latest
```

### Usage
```
bbrf services where tools is armada -p hbo | armada-nmap
```

### Input format
```
127.0.0.1:80
127.0.0.1:8080
10.0.0.1:80
10.0.0.1:443
10.0.0.1:3000
```