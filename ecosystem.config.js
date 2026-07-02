// PM2：blog-server-go 四微服务（gateway :8000 + user/blog/rpg）
// 启动：pm2 start ecosystem.config.js --env production
// 日常发布：deploy/pm2/deploy.ps1（本地交叉编译 + tar 上传 + pm2 reload）

const shared = {
  cwd: './',
  interpreter: 'none',
  exec_mode: 'fork',
  instances: 1,
  autorestart: true,
  kill_timeout: 8000,
  listen_timeout: 8000,
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
      name: 'gateway',
      script: './bin/gateway',
      max_memory_restart: '120M',
      error_file: '../logs/gateway-err.log',
      out_file: '../logs/gateway-out.log',
      env_production: {
        CONFIG_PATH: './configs/gateway.yaml',
        GOMEMLIMIT: '60MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'user',
      script: './bin/user',
      max_memory_restart: '140M',
      error_file: '../logs/user-err.log',
      out_file: '../logs/user-out.log',
      env_production: {
        CONFIG_PATH: './configs/user.yaml',
        GOMEMLIMIT: '70MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'blog',
      script: './bin/blog',
      max_memory_restart: '160M',
      error_file: '../logs/blog-err.log',
      out_file: '../logs/blog-out.log',
      env_production: {
        CONFIG_PATH: './configs/blog.yaml',
        GOMEMLIMIT: '90MiB',
        GOGC: '50',
      },
    },
    {
      ...shared,
      name: 'rpg',
      script: './bin/rpg',
      max_memory_restart: '160M',
      error_file: '../logs/rpg-err.log',
      out_file: '../logs/rpg-out.log',
      env_production: {
        CONFIG_PATH: './configs/rpg.yaml',
        GOMEMLIMIT: '90MiB',
        GOGC: '50',
      },
    },
  ],
};
