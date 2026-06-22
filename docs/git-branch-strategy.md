# Git 分支策略

PlantX 采用 **Git Flow + Trunk-Based 混合模型**：

- 日常开发以 `main` 为主干，通过短生命周期的 feature/fix 分支合并。
- 发布周期较长时，从 `main` 切出 `release/vx.y.z` 分支进行冻结测试。

## 分支命名

- `feature/<issue-id>-description`
- `fix/<issue-id>-description`
- `release/vx.y.z`
- `hotfix/<description>`

## 保护规则

- `main` 和 `release/*` 必须：
  - 至少 1 个 PR Review
  - CI 全部通过
  - 分支必须是最新
- 禁止直接推送。

## Commit 规范

- 使用 Conventional Commits：`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`。
- CI 通过 commitlint 校验。

## 发布流程

1. 创建 `release/vx.y.z` 分支。
2. 跑完整集成测试（Docker Compose、K8s、二进制）。
3. 打 tag `vx.y.z`。
4. 将 release 分支合并回 `main`。

## Hotfix 流程

1. 从最新 tag 切出 `hotfix/<description>`。
2. 修复后 tag `vx.y.z+1`。
3. 合并到 `main` 和活动中的 `release/*`。

## Monorepo 版本

- 使用 changeset 对 `kit-go`、`kit-cli`、`kit-ui`、平台服务、业务服务独立版本管理。
- CI 在 tag push 时构建并发布所有容器镜像到仓库。

## PR Merge 规则

- feature/fix：squash merge，保持历史简洁。
- release/hotfix：merge commit，保留分支信息。
- 必须配置 CODEOWNERS。
