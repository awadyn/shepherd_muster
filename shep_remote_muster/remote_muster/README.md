In this directory are objects that represent a remote muster component running on a remote (relative to the shepherd) server or worker node alongside a particular application.
The remote muster is the preliminary and core specialization of a muster component. The core muster is specialized to 1) 'attach to' a logging framework running natively on the remote node and 2) have access to native control primitives on the remote node.
* ```remote_muster.go```: functionality to support server and client threads that coordinate with client and server threads of the mirror local muster to manage log and control synchronization 
* ```muster_remote.go```: functionality to support remote muster threads that coordinate log recording and control application to different sheep on the remote node


The core muster functionality remains common between local and remote muster specializations and includes the 'sheep' objects - computational resources that a muster must manage - and their attendant log and control objects.

The core muster specialization remains common as well and includes sheep specializations: log metrics and buffer size specifications and a set of control settings chosen for optimization. 
