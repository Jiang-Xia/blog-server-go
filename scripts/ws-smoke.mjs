#!/usr/bin/env node
/**
 * Plan 08 WebSocket 冒烟：HTTP since + 原生 WS 心跳 + dev 推送 + Stream 发布
 * 用法：先启动 make dev，再 node scripts/ws-smoke.mjs
 */
const BASE = process.env.BASE_URL || 'http://localhost:5000';
const API = `${BASE}/api/v1`;

async function loginToken() {
  const { execSync } = await import('node:child_process');
  const goDir = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1');
  try {
    const out = execSync('go run scripts/dev_login.go --token-only', {
      cwd: goDir.replace(/\//g, '\\').replace(/\\$/, ''),
      encoding: 'utf8',
      env: { ...process.env, CONFIG_PATH: 'configs/monolith.yaml' },
    }).trim();
    return out.split('\n').pop().trim();
  }
  catch (e) {
    console.error('dev_login 失败，请确认 blog-server-go 可运行且 MySQL/Redis 可用');
    throw e;
  }
}

async function httpGet(path, token) {
  const res = await fetch(`${API}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const body = await res.json();
  if (!res.ok || body.success === false) {
    throw new Error(`GET ${path} failed: ${JSON.stringify(body)}`);
  }
  return body.data;
}

async function httpPost(path, token, query = '') {
  const res = await fetch(`${API}${path}${query}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  const body = await res.json();
  if (!res.ok || body.success === false) {
    throw new Error(`POST ${path} failed: ${JSON.stringify(body)}`);
  }
  return body.data;
}

function wsSmoke(token) {
  return new Promise((resolve, reject) => {
    const wsUrl = BASE.replace(/^http/i, 'ws') + `/realtime?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(wsUrl);
    const timer = setTimeout(() => {
      ws.close();
      reject(new Error('WS 超时'));
    }, 15000);

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: 'ping' }));
    };

    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data);
      if (msg.type === 'pong') {
        clearTimeout(timer);
        ws.close();
        resolve(true);
      }
    };

    ws.onerror = (e) => {
      clearTimeout(timer);
      reject(e);
    };
  });
}

async function main() {
  console.log('1. 登录获取 token…');
  const token = await loginToken();

  console.log('2. GET /notification/since?seq=0');
  const since = await httpGet('/notification/since?seq=0', token);
  console.log(`   since 返回 ${Array.isArray(since) ? since.length : 0} 条`);

  console.log('3. WS 连接 + 应用层 ping/pong');
  await wsSmoke(token);
  console.log('   pong 收到');

  console.log('4. POST /dev/ws-push');
  await httpPost('/dev/ws-push', token, '?type=smokeTest');
  console.log('   Hub 直推 OK');

  console.log('5. POST /dev/ws-push-redis');
  await httpPost('/dev/ws-push-redis', token);
  console.log('   Redis pub/sub 推送 OK');

  console.log('6. POST /dev/event-publish');
  await httpPost('/dev/event-publish', token);
  console.log('   Stream 发布 OK');

  console.log('\n✅ Plan 08 WS 冒烟通过');
}

main().catch((err) => {
  console.error('\n❌ 冒烟失败:', err.message || err);
  process.exit(1);
});
