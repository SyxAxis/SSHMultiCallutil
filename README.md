# SSHMultiCallutil
Simple util to make multi command calls over SSH using priv key

More practice playing with Go. This is very simple util that reads a JSON config file that has some key connection details plus some command lines. It connection, runs each command line. When the command lines are running they're feeding back through local stdErr and stdOut in realtime, so should be "pipeable" through anything that can read those streams locally.

The idea was to have a util that can slap down a script onto a Unix box, run it and remove it afterwards, then all the temporary unix scripts can be stored in config files in repos like Git, maintained off host and easier to maintain than constantly logging into the remote host to edit scripts. I'll do that in the next stage when I get some time. 

Right now this is just a piece of educutional code for example purposes only. It's perfectly safe SSL code so with the proper tweaking could be make production ready.
