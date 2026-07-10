// PM2：blog-server-go 单体 monolith（:8000，生产对外端口）
// 实例名：BlogGo_Monolith
//
// 生产 cwd 固定为 $DEPLOY_REMOTE_DIR/current（软链），切换 release 后 pm2 reload 即可加载新二进制。
// 本地勿用 pm2 跑本文件；开发用 make dev / .\scripts\dev.ps1

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
      name: 'BlogGo_Monolith',
      script: './bin/monolith',
      max_memory_restart: '320M',
      error_file: `${logDir}/monolith-err.log`,
      out_file: `${logDir}/monolith-out.log`,
      env_production: {
        CONFIG_PATH: './configs/monolith.yaml',
        DEPLOY_REMOTE_DIR: deployRoot,
        GOMEMLIMIT: '280MiB',
        GOGC: '50',
      },
    },
  ],
};
