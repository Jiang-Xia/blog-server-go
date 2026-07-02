// PM2：blog-server-go 四微服务（gateway :8000 + user/blog/rpg）
// 实例名：BlogGo_{Gateway|User|Blog|Rpg}
//
// 生产 cwd 固定为 $DEPLOY_REMOTE_DIR/current（软链），切换 release 后 pm2 reload 即可加载新二进制，无需 delete。
// 本地勿用 pm2 跑本文件；开发用 make dev-all。

const deployRoot = process.env.DEPLOY_REMOTE_DIR || '';
const appCwd = deployRoot ? `${deployRoot}/current` : '.';
const logDir = deployRoot ? `${deployRoot}/logs` : '../logs';

const shared = {
  cwd: appCwd,
  interpreter: 'none',
  exec_mode: 'fork',
  instances: 1,
  autorestart: true,
  kill_timeout: 8000,
  listen_timeout: 10000,
  merge_logs: true,
  log_date_format: 'YYYY-MM-DD HH:mm:ss',
  min_uptime: '30s',
  max_restarts: 30,
  restart_delay: 5,
  watch: false,
};

module.exports = {
  apps: [
    {
      ...shared,
      name: 'BlogGo_Gateway',
      script: './bin/gateway',
      max_memory_restart: '120M',
      error_file: `${logDir}/gateway-err.log`,
      out_file: `${logDir}/gateway-out.log`,
      env_production: {
        CONFIG_PATH: './configs/gateway.yaml',
        DEPLOY_REMOTE_DIR: deployRoot,
        GOMEMLIMIT: '60MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'BlogGo_User',
      script: './bin/user',
      max_memory_restart: '140M',
      error_file: `${logDir}/user-err.log`,
      out_file: `${logDir}/user-out.log`,
      env_production: {
        CONFIG_PATH: './configs/user.yaml',
        DEPLOY_REMOTE_DIR: deployRoot,
        GOMEMLIMIT: '70MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'BlogGo_Blog',
      script: './bin/blog',
      max_memory_restart: '160M',
      error_file: `${logDir}/blog-err.log`,
      out_file: `${logDir}/blog-out.log`,
      env_production: {
        CONFIG_PATH: './configs/blog.yaml',
        DEPLOY_REMOTE_DIR: deployRoot,
        GOMEMLIMIT: '90MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'BlogGo_Rpg',
      script: './bin/rpg',
      max_memory_restart: '160M',
      error_file: `${logDir}/rpg-err.log`,
      out_file: `${logDir}/rpg-out.log`,
      env_production: {
        CONFIG_PATH: './configs/rpg.yaml',
        DEPLOY_REMOTE_DIR: deployRoot,
        GOMEMLIMIT: '90MiB',
        GOGC: '50',
      },
    },
  ],
};
