The purpose of this small command line program written in Golang is to be able to easily replace a tag in an env file on a remote environment.
Part of deploying my code on my lab at work is replacing a tag generated through the CD pipeline. To do that, the team needs to update the .env file. The .env file contains key value pairs in the following format

```
KEY=<IP>/<UNIQUE_ID>:<UNIQUE_TAG>
```

This is simple using Vim for me. However, junior members of the team do not know Vim. Rather than to risk them making mistakes, to ease them with updating the tag and to delegate the responsibility of deployment from me to other team members, I created this program to facilitate the process