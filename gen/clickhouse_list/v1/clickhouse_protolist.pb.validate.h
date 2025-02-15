// Code generated by protoc-gen-validate
// source: clickhouse_list/v1/clickhouse_protolist.proto
// DO NOT EDIT!!!

#pragma once

#include <algorithm>
#include <string>
#include <sstream>
#include <unordered_set>
#include <vector>

#include "validate/validate.h"
#include "clickhouse_list/v1/clickhouse_protolist.pb.h"


namespace clickhouse_protolist {
namespace v1 {

using std::string;


extern bool Validate(const ::clickhouse_protolist::v1::Record& m, pgv::ValidationMsg* err);

extern bool Validate(const ::clickhouse_protolist::v1::Envelope& m, pgv::ValidationMsg* err);


} // namespace
} // namespace


#define X_CLICKHOUSE_PROTOLIST_V1_CLICKHOUSE_PROTOLIST(X) \
X(::clickhouse_protolist::v1::Record) \
X(::clickhouse_protolist::v1::Envelope) \

