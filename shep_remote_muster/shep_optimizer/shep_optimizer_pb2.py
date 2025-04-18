# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: shep_optimizer/shep_optimizer.proto
# Protobuf Python Version: 5.29.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    0,
    '',
    'shep_optimizer/shep_optimizer.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n#shep_optimizer/shep_optimizer.proto\x12\x08shepherd\")\n\x0c\x43ontrolEntry\x12\x0c\n\x04knob\x18\x01 \x01(\t\x12\x0b\n\x03val\x18\x02 \x01(\x04\"&\n\x0bRewardEntry\x12\n\n\x02id\x18\x01 \x01(\t\x12\x0b\n\x03val\x18\x02 \x01(\x02\")\n\x15StartOptimizerRequest\x12\x10\n\x08n_trials\x18\x01 \x01(\r\"#\n\x13StartOptimizerReply\x12\x0c\n\x04\x64one\x18\x01 \x01(\x08\"\x16\n\x14StopOptimizerRequest\"\"\n\x12StopOptimizerReply\x12\x0c\n\x04\x64one\x18\x01 \x01(\x08\"8\n\x0fOptimizeRequest\x12%\n\x05\x63trls\x18\x01 \x03(\x0b\x32\x16.shepherd.ControlEntry\"E\n\rOptimizeReply\x12\x0c\n\x04\x64one\x18\x01 \x01(\x08\x12&\n\x07rewards\x18\x02 \x03(\x0b\x32\x15.shepherd.RewardEntry2\xb0\x01\n\rSetupOptimize\x12P\n\x0eStartOptimizer\x12\x1f.shepherd.StartOptimizerRequest\x1a\x1d.shepherd.StartOptimizerReply\x12M\n\rStopOptimizer\x12\x1e.shepherd.StopOptimizerRequest\x1a\x1c.shepherd.StopOptimizerReply2S\n\x08Optimize\x12G\n\x11\x45valuateOptimizer\x12\x19.shepherd.OptimizeRequest\x1a\x17.shepherd.OptimizeReplyB&Z$github.com/awadyn/shep_remote_musterb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'shep_optimizer.shep_optimizer_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z$github.com/awadyn/shep_remote_muster'
  _globals['_CONTROLENTRY']._serialized_start=49
  _globals['_CONTROLENTRY']._serialized_end=90
  _globals['_REWARDENTRY']._serialized_start=92
  _globals['_REWARDENTRY']._serialized_end=130
  _globals['_STARTOPTIMIZERREQUEST']._serialized_start=132
  _globals['_STARTOPTIMIZERREQUEST']._serialized_end=173
  _globals['_STARTOPTIMIZERREPLY']._serialized_start=175
  _globals['_STARTOPTIMIZERREPLY']._serialized_end=210
  _globals['_STOPOPTIMIZERREQUEST']._serialized_start=212
  _globals['_STOPOPTIMIZERREQUEST']._serialized_end=234
  _globals['_STOPOPTIMIZERREPLY']._serialized_start=236
  _globals['_STOPOPTIMIZERREPLY']._serialized_end=270
  _globals['_OPTIMIZEREQUEST']._serialized_start=272
  _globals['_OPTIMIZEREQUEST']._serialized_end=328
  _globals['_OPTIMIZEREPLY']._serialized_start=330
  _globals['_OPTIMIZEREPLY']._serialized_end=399
  _globals['_SETUPOPTIMIZE']._serialized_start=402
  _globals['_SETUPOPTIMIZE']._serialized_end=578
  _globals['_OPTIMIZE']._serialized_start=580
  _globals['_OPTIMIZE']._serialized_end=663
# @@protoc_insertion_point(module_scope)
