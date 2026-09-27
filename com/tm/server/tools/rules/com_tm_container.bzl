"""Macro build binary + OCI image (cấu trúc theo project thor).

Mỗi macro sinh cùng một bộ target:

    //path/to/app:<name>          binary chạy local (bazel run)
    //path/to/app:<name>_image    oci_image
    //path/to/app:<name>_docker   load image vào Docker local
    //path/to/app:<name>_push     push image (chỉ khi truyền `repository`)

Image tag: com.tm.go.<image_name>:v1.0.0 (image_name mặc định = name).

Image chỉ build cho Linux — chọn kiến trúc máy chạy container:

    bazel run --config=linux-arm64 //services/core/cmd/server:server_docker   # Apple Silicon
    bazel run --config=linux-amd64 //services/core/cmd/server:server_docker   # server x86

Không --config (vd `bazel build //...` trên macOS) → target image bị bỏ qua.
"""

load("@bazel_skylib//rules:copy_file.bzl", "copy_file")
load("@rules_go//go:def.bzl", "go_binary")
load("@rules_oci//oci:defs.bzl", "oci_image", "oci_load", "oci_push")
load("@tar.bzl", "tar")

DEFAULT_IMAGE_TAG = "v1.0.0"

# User nonroot có sẵn trong distroless (:nonroot); ghi rõ để không phụ thuộc tag base.
_NONROOT_USER = "65532:65532"

# Image chỉ có nghĩa trên Linux; build cho nền tảng khác → Bazel bỏ qua thay vì báo lỗi.
_LINUX_ONLY = ["@platforms//os:linux"]

def _container_targets(
        name,
        base,
        repo_tag,
        tar_srcs = [],
        layers = [],
        entrypoint = None,
        cmd = None,
        env = None,
        workdir = None,
        exposed_ports = [],
        user = _NONROOT_USER,
        repository = None,
        image_tag = DEFAULT_IMAGE_TAG,
        visibility = ["//visibility:public"]):
    """Sinh image / docker load / push — phần chung cho mọi ngôn ngữ.

    `tar_srcs` được đóng thành một layer `<name>_tar`; `layers` là các layer tar
    dựng sẵn, thêm sau layer đó.
    """

    tars = list(layers)
    if tar_srcs:
        tar(
            name = name + "_tar",
            srcs = tar_srcs,
            out = name + "_layer.tar",
            target_compatible_with = _LINUX_ONLY,
            visibility = visibility,
        )
        tars = [":" + name + "_tar"] + tars

    oci_image(
        name = name + "_image",
        base = base,
        cmd = cmd,
        entrypoint = entrypoint,
        env = env,
        exposed_ports = exposed_ports,
        tars = tars,
        target_compatible_with = _LINUX_ONLY,
        user = user,
        visibility = visibility,
        workdir = workdir,
    )

    oci_load(
        name = name + "_docker",
        image = ":" + name + "_image",
        repo_tags = ["%s:%s" % (repo_tag, image_tag)],
        target_compatible_with = _LINUX_ONLY,
        visibility = visibility,
    )

    if repository:
        oci_push(
            name = name + "_push",
            image = ":" + name + "_image",
            remote_tags = [image_tag],
            repository = repository,
            target_compatible_with = _LINUX_ONLY,
            visibility = visibility,
        )

# =============================================================================
# GO
# =============================================================================

def com_tm_go_image(
        name,
        embed,
        image_name = None,
        data = [],
        args = [],
        exposed_ports = [],
        env = None,
        repository = None,
        image_tag = DEFAULT_IMAGE_TAG,
        visibility = ["//visibility:public"]):
    """go_binary (static, CGO off) + OCI image trên distroless static.

    Gazelle sinh target này thay cho go_binary (map_kind trong BUILD.bazel gốc).

    Args:
        name: tên gốc cho mọi target.
        embed: go_library chứa `package main` (do gazelle sinh).
        image_name: tên trong tag `com.tm.go.<image_name>`; mặc định = name.
        data: file runtime, cũng được bake vào image.
        args: tham số mặc định khi chạy local.
        exposed_ports: port expose trong container, vd ["8080/tcp"].
        env: biến môi trường cho container.
        repository: registry cho `<name>_push`; bỏ trống để không sinh target push.
        image_tag: tag image.
        visibility: visibility của target sinh ra.
    """

    go_binary(
        name = name,
        args = args,
        data = data,
        embed = embed,
        pure = "on",
        static = "on",
        visibility = visibility,
    )

    # rules_go đặt binary ở <name>_/<name>; copy ra path ổn định cho entrypoint.
    copy_file(
        name = name + "_bin",
        src = ":" + name,
        out = "bin/" + name,
        is_executable = True,
        target_compatible_with = _LINUX_ONLY,
    )

    # tar giữ đường dẫn package → trong image: /<package>/bin/<name>
    _container_targets(
        name = name,
        base = "@distroless_static",
        entrypoint = ["/%s/bin/%s" % (native.package_name(), name)],
        env = env,
        exposed_ports = exposed_ports,
        image_tag = image_tag,
        repo_tag = "com.tm.go.%s" % (image_name or name),
        repository = repository,
        tar_srcs = [":" + name + "_bin"] + data,
        visibility = visibility,
    )
