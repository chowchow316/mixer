package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "check.go",
        "main.go",
        "quota.go",
        "report.go",
        "util.go",
    ],
    visibility = ["//visibility:private"],
    deps = [
        "@com_github_golang_glog//:go_default_library",
        "@com_github_golang_protobuf//ptypes:go_default_library",
        "@com_github_golang_protobuf//ptypes/timestamp:go_default_library",
        "@com_github_istio_api//:go_default_library",
        "@com_github_spf13_cobra//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
    ],
)

go_test(
    name = "util_test",
    size = "small",
    srcs = ["util_test.go"],
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_binary(
    name = "mixc",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)