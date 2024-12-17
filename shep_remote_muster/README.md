In this directory are the main components of MustHerd:
1. The shepherd (```shepherd/```) subcomponent consists of the centralized component that coordinates execution data with learning and optimization mechanisms.
2. The remote muster subcomponent (```remote_muster/```) consists of the one-or-more distributed muster components that i) record and send execution data to the shepherd and ii) receive and apply control decisions made by the shepherd.
3. The messaging protocol between shepherd and remote muster components (```shep_remote_muster/```) which consists of grpc calls and message definitions that support core MustHerd functionality: log and control coordination between the centralized shepherd and remote musters.
4. The messaging protocol between shepherd and optimizer components (```shep_optimizer/```) which consists of grpc calls and message definitions that support specialized MustHerd functionality: data and decision coordination between the shepherd and a specialized optimization mechanism.

By virtue of local and remote muster mirroring that MustHerd relies on, several core subcomponents are common in shepherd and remote muster deployments (```defs.go, muster.go, <specialized>_muster.go```).
