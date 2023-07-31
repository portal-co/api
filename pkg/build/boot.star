def base_build_rule(name,cfg,ctx=global_ctx):
    def go(sub_ctx):
        c = cfg(sub_ctx.cfg)
        d = {}
        for k, v in c.deps:
            d[k] = sub_ctx.target(v.target,lambda pkg_cfg: v.transition(sub_ctx.cfg))["default"]
        return {
            "default": sub_ctx.run(d,c.cmd,c.outs)
        }
    ctx.bind(name, go)

CPUS = {"x86_64": "amd64", "aarch64": "arm64", "riscv64": "riscv64"}
RCPUS = {"amd64": "x86_64", "arm64": "aarch64", "riscv64": "riscv64"}
OS = {
    "linux": ["linux"],
    "macos": ["macos", "darwin"],
    "windows": ["windows"],
    "unknown": None
}
GLIBCS = [
    "2.17",
    "2.18",
    "2.19",
    "2.22",
    "2.23",
    "2.24",
    "2.25",
    "2.26",
    "2.27",
    "2.28",
    "2.29",
    "2.30",
    "2.31",
    "2.32",
    "2.33",
    "2.34",
]
LIBCS = {
    "linux": ["musl"] + ["gnu.{}".format(glibc) for glibc in _GLIBCS],
    "macos": ["inbuilt"],
    "windows": ["gnu","msvc"],
    "unknown": ["none"]
}

def to_go(cfg):
    if "target" not in cfg:
        return None
    tn = cfg["target"].split('-')
    goos = tn[1]
    gocpu = CPUS[tn[0]]
    return struct(goos = goos, goarch = gocpu)

def from_go(goos,goarch):
    r = RCPUS[goarch]
    return f"{r}-{goos}"

def go_env(cfg):
    t = to_go(cfg)
    if t == None:
        return []
    return ["GOOS=" + t.goos, "GOARCH=" + t.goarch]

def boot_go_binary(name,src,goname,gotarget,ctx=global_ctx):
    base_build_rule(name = name, ctx = ctx, cfg = lambda cfg: struct(cmd = ["/usr/bin/env"] + go_env(cfg) + ["sh","-c",f"cd ./src;go build -buildmode {gotarget} -o ../out {goname}"], outs = ["out"], deps = {"src": struct(target = src, transition: lambda x: x)}))

def to_host(cfg):
    return {}