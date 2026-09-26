"""Macro build chương trình Go kèm OCI image.

Một lệnh gọi com_tm_go_image(name = "x", ...) sinh:

    :x          go_binary chạy local           bazel run //path:x
    :x_image    oci_image (distroless)          bazel build --config=linux-arm64 //path:x_image
    :x_docker   load image vào Docker local     bazel run --config=linux-arm64 //path:x_docker
    :x_push     push image (chỉ khi có `repository`)

Tag image: com.tm.go.<name>:<image_tag>

Gazelle map `go_binary` sang macro này (BUILD.bazel gốc), nên chương trình mới
có sẵn các target trên sau `bazel run //:gazelle`.
"""

load("@bazel_skylib//rules:copy_file.bzl", "copy_file")
load("@rules_go//go:def.bzl", "go_binary")
load("@rules_oci//oci:defs.bzl", "oci_image", "oci_load", "oci_push")
load("@tar.bzl", "tar")

DEFAULT_IMAGE_TAG = "v1.0.0"

def com_tm_go_image(
        name,
        embed,
        package_name = None,
        data = [],
        args = [],
        exposed_ports = [],
        env = None,
        repository = None,
        image_tag = DEFAULT_IMAGE_TAG,
        visibility = ["//visibility:public"],
        **kwargs):
    """go_binary tĩnh (CGO tắt) + OCI image trên distroless.

    Args:
        name: tên gốc cho mọi target.
        embed: go_library chứa `package main` (gazelle sinh).
        package_name: package Bazel chứa binary; mặc định là package hiện tại.
        data: file runtime, cũng được đưa vào image.
        args: tham số mặc định khi chạy local.
        exposed_ports: port expose trong container, vd ["8080/tcp"].
        env: biến môi trường của container.
        repository: registry cho `<name>_push`; bỏ trống thì không sinh target push.
        image_tag: tag image.
        visibility: visibility của các target sinh ra.
        **kwargs: thuộc tính khác chuyển thẳng cho go_binary.
    """
    pkg = package_name if package_name != None else native.package_name()

    go_binary(
        name = name,
        args = args,
        data = data,
        embed = embed,
        pure = "on",
        static = "on",
        visibility = visibility,
        **kwargs
    )

    # rules_go đặt binary ở <name>_/<name>; copy ra đường dẫn ổn định cho entrypoint.
    copy_file(
        name = name + "_bin",
        src = ":" + name,
        out = "bin/" + name,
        is_executable = True,
    )

    tar(
        name = name + "_tar",
        srcs = [":" + name + "_bin"] + data,
        out = name + "_layer.tar",
    )

    oci_image(
        name = name + "_image",
        base = "@distroless_base",
        entrypoint = ["/%s/bin/%s" % (pkg, name)],
        env = env,
        exposed_ports = exposed_ports,
        tars = [":" + name + "_tar"],
        visibility = visibility,
    )

    oci_load(
        name = name + "_docker",
        image = ":" + name + "_image",
        repo_tags = ["com.tm.go.%s:%s" % (name, image_tag)],
        visibility = visibility,
    )

    if repository:
        oci_push(
            name = name + "_push",
            image = ":" + name + "_image",
            remote_tags = [image_tag],
            repository = repository,
            visibility = visibility,
        )
