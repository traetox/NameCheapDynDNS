NameCheapDynDNS
===============

Namecheap dynamic DNS record updater client.

NameCheapDynDNS is used to monitor and update multiple NameCheap dynamic
DNS entries.  It supports NAT by reaching out to a remote URL to determine
the IP of the system to update.

Multiple domains can be updated at once via the settings.conf file.


And example settings.conf file is:


```
[global]
  UpdateInterval=10 #check every 10 minutes

[DynDomain "gitserver"] #a named entry, name can be anything
  Host="git" #the subdomain A record host you wish to update
  Domain="mysuperprivatedomain.com" #your domain
  Key="AABBCCDDEEFFGGHHIIJJKKI"#dynamic IP key provided by NameCheap


[DynDomain "fileserver"]
  Host="home"
  Domain="iownthisdomain.net"
  Key="BBCCDDEEFFFHHIIKJJKJLKSJDLKJ"
```
