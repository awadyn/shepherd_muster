In this directory are objects that represent a muster component running on a remote server or worker node alongside a particular application.
The remote muster is a preliminary and core specialization of a muster component. A core muster is specialized to:
1. 'attach to' a logging framework running natively on the remote node and
2. have access to control primitives on the remote node.

The core muster functionality remains common between local and remote muster specializations and includes the 'sheep' objects - computational resources that a muster must manage - and their attendant log and control objects.

The core muster specialization remains common as well and includes sheep specializations: log metrics and buffer size specifications and a set of control settings chosen for optimization. 
