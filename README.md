# infrakit.dind

- It's for testing 

```
{
  "ID": "swarm-dind",
  "Properties": {
    "Allocation": {
      "Size": 2
    },
    "Instance": {
      "Plugin": "instance-dind",
      "Properties": {
		 "Name": "worker",
		 "HostName" : "worker"
      }
    },
    "Flavor": {
      "Plugin": "flavor-vanilla",
      "Properties": {
          "Init": [
              "docker swarm join --token SWMTKN-1-3cfnsnumc4ptz1ame7rac2dgq4atklr9nza6amux438jkd02g9-csb58zcltq94m5uf7q2im4dvi 192.168.65.2:2377"
          ]
      }
    }
  }
}

```


