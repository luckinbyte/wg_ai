# 热更新使用指南

## 快速开始

### 1. 编译插件

```bash
make plugin-role    # 编译角色模块
make plugin-item    # 编译物品模块
make plugins        # 编译所有插件
```

### 2. 启动服务

```bash
make build
./bin/game -config ./config/game.yaml
```

### 3. 热更插件

**方式一: HTTP API**

```bash
# 编译新版本
make plugin-role

# 调用热更接口
curl -X POST http://localhost:8081/admin/hotreload \
  -H "Content-Type: application/json" \
  -d '{"module": "role", "path": "./plugins/role.so"}'
```

**方式二: 文件监听 (自动)**

配置 `plugin.watch: true` 后，直接替换 `plugins/` 目录下的 `.so` 文件即可自动触发热更。

### 4. 查看已加载插件

```bash
curl http://localhost:8081/admin/plugins
```

### 5. 健康检查

```bash
curl http://localhost:8081/admin/health
```

## 注意事项

1. **编译一致性**: 主程序和插件必须使用相同的 Go 版本编译
2. **内存泄漏**: 旧版本插件无法卸载，多次热更会累积内存
3. **建议在低峰期执行热更操作**
4. **热更前先在测试环境验证**

## 开发新插件

1. 在 `plugin/` 目录创建新模块:

```go
// plugin/mymodule/mymodule.go
package main

import baseplugin "github.com/yourorg/wg_ai/plugin"

type MyLogic struct{}

func (l *MyLogic) Name() string { return "mymodule" }

func (l *MyLogic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
    // 实现逻辑
    return baseplugin.Success(nil), nil
}

var MyModule baseplugin.LogicModule = &MyLogic{}
```

2. 添加路由配置 `config/routes.json`:

```json
{"msg_id": 3001, "module": "mymodule", "method": "do_something"}
```

3. 添加编译目标到 Makefile:

```makefile
plugin-mymodule:
	go build -buildmode=plugin -o plugins/mymodule.so ./plugin/mymodule
```
