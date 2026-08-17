import { defineConfig } from 'vitepress'

const repo = 'https://github.com/mouse233/MySekaiMapper'

const enSidebar = [
  {
    text: 'Guide',
    items: [
      { text: 'Introduction', link: '/guide/introduction' },
      { text: 'Quick Start', link: '/guide/quickstart' },
      { text: 'Upload API', link: '/guide/upload-api' },
      { text: 'Reqable Report Server', link: '/guide/report-server' },
      { text: 'Push Mechanism', link: '/guide/push' },
      { text: 'Static File Server', link: '/guide/static-server' },
      { text: 'Player Routing', link: '/guide/routing' },
      { text: 'CLI Reference', link: '/guide/cli' }
    ]
  },
  {
    text: 'Project',
    items: [
      { text: 'Directory Structure', link: '/reference/structure' },
      { text: 'FAQ', link: '/faq' },
      { text: 'License', link: '/license' }
    ]
  }
]

const zhSidebar = [
  {
    text: '指南',
    items: [
      { text: '项目介绍', link: '/zh-CN/guide/introduction' },
      { text: '快速上手', link: '/zh-CN/guide/quickstart' },
      { text: '上传接口', link: '/zh-CN/guide/upload-api' },
      { text: 'Reqable 上报服务器', link: '/zh-CN/guide/report-server' },
      { text: '推送机制', link: '/zh-CN/guide/push' },
      { text: '静态文件服务器', link: '/zh-CN/guide/static-server' },
      { text: '玩家推送路由', link: '/zh-CN/guide/routing' },
      { text: '命令行工具', link: '/zh-CN/guide/cli' }
    ]
  },
  {
    text: '项目',
    items: [
      { text: '目录结构', link: '/zh-CN/reference/structure' },
      { text: '常见问题', link: '/zh-CN/faq' },
      { text: '许可证', link: '/zh-CN/license' }
    ]
  }
]

const zhTWSidebar = [
  {
    text: '指南',
    items: [
      { text: '專案介紹', link: '/zh-TW/guide/introduction' },
      { text: '快速上手', link: '/zh-TW/guide/quickstart' },
      { text: '上傳介面', link: '/zh-TW/guide/upload-api' },
      { text: 'Reqable 上報伺服器', link: '/zh-TW/guide/report-server' },
      { text: '推播機制', link: '/zh-TW/guide/push' },
      { text: '靜態檔案伺服器', link: '/zh-TW/guide/static-server' },
      { text: '玩家推播路由', link: '/zh-TW/guide/routing' },
      { text: '命令列工具', link: '/zh-TW/guide/cli' }
    ]
  },
  {
    text: '項目',
    items: [
      { text: '目錄結構', link: '/zh-TW/reference/structure' },
      { text: '常見問題', link: '/zh-TW/faq' },
      { text: '授權條款', link: '/zh-TW/license' }
    ]
  }
]

const jaJPSidebar = [
  {
    text: 'ガイド',
    items: [
      { text: 'プロジェクト紹介', link: '/ja-JP/guide/introduction' },
      { text: 'クイックスタート', link: '/ja-JP/guide/quickstart' },
      { text: 'アップロードAPI', link: '/ja-JP/guide/upload-api' },
      { text: 'Reqable レポートサーバー', link: '/ja-JP/guide/report-server' },
      { text: 'プッシュの仕組み', link: '/ja-JP/guide/push' },
      { text: '静的ファイルサーバー', link: '/ja-JP/guide/static-server' },
      { text: 'プレイヤープッシュルーティング', link: '/ja-JP/guide/routing' },
      { text: 'コマンドラインツール', link: '/ja-JP/guide/cli' }
    ]
  },
  {
    text: 'プロジェクト',
    items: [
      { text: 'ディレクトリ構成', link: '/ja-JP/reference/structure' },
      { text: 'よくある質問', link: '/ja-JP/faq' },
      { text: 'ライセンス', link: '/ja-JP/license' }
    ]
  }
]

const koKRSidebar = [
  {
    text: '가이드',
    items: [
      { text: '프로젝트 소개', link: '/ko-KR/guide/introduction' },
      { text: '빠른 시작', link: '/ko-KR/guide/quickstart' },
      { text: '업로드 API', link: '/ko-KR/guide/upload-api' },
      { text: 'Reqable 보고서 서버', link: '/ko-KR/guide/report-server' },
      { text: '푸시 메커니즘', link: '/ko-KR/guide/push' },
      { text: '정적 파일 서버', link: '/ko-KR/guide/static-server' },
      { text: '플레이어 푸시 라우팅', link: '/ko-KR/guide/routing' },
      { text: '명령줄 도구', link: '/ko-KR/guide/cli' }
    ]
  },
  {
    text: '프로젝트',
    items: [
      { text: '디렉터리 구조', link: '/ko-KR/reference/structure' },
      { text: '자주 묻는 질문', link: '/ko-KR/faq' },
      { text: '라이선스', link: '/ko-KR/license' }
    ]
  }
]

export default defineConfig({
  // GitHub Pages 项目站点必须设置 base
  base: '/MySekaiMapper/',
  cleanUrls: true,
  lastUpdated: true,
  markdown: {
    lineNumbers: true
  },
  // 仅对默认语言(英文)首页注入"按系统语言自动跳转"脚本:
  // 在页面渲染前读取 navigator.language,若系统语言为中文/日文/韩文则重定向到对应语言首页。
  // 只匹配主站根路径,用户已访问的特定语言页面不受影响。
  transformHtml(code, _id, ctx) {
    if (ctx.page === 'index.md') {
      const script = `<script>
(function () {
  var base = '/MySekaiMapper/';
  var path = location.pathname.replace(/\\/+$/, '');
  if (path !== base.replace(/\\/+$/, '')) return;
  var lang = (navigator.language || 'en').toLowerCase();
  var map = [
    ['zh-tw', 'zh-TW/'],
    ['zh-hk', 'zh-TW/'],
    ['zh-cn', 'zh-CN/'],
    ['zh', 'zh-CN/'],
    ['ja', 'ja-JP/'],
    ['ko', 'ko-KR/']
  ];
  for (var i = 0; i < map.length; i++) {
    if (lang.indexOf(map[i][0]) === 0) {
      location.replace(base + map[i][1]);
      return;
    }
  }
})();
</script>`
      return code.replace('</head>', script + '\n</head>')
    }
    return code
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      title: 'MySekaiMapper',
      description:
        'A resource-gathering point map generator and auto-notifier for MySekai in Project SEKAI',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/introduction', activeMatch: '/guide/' },
          { text: 'FAQ', link: '/faq' },
          { text: 'License', link: '/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: enSidebar,
        outline: { label: 'On this page', level: [2, 3] },
        docFooter: { prev: 'Previous page', next: 'Next page' },
        lastUpdated: { text: 'Last updated' },
        returnToTopLabel: 'Back to top',
        sidebarMenuLabel: 'Menu',
        darkModeSwitchLabel: 'Appearance',
        lightModeSwitchTitle: 'Switch to light mode',
        darkModeSwitchTitle: 'Switch to dark mode',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: 'Edit this page on GitHub'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    },
    'zh-CN': {
      label: '简体中文',
      lang: 'zh-CN',
      title: 'MySekaiMapper',
      description: '《世界计划 多彩舞台》MySekai 采集点地图生成与自动推送工具',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh-CN/guide/introduction', activeMatch: '/zh-CN/guide/' },
          { text: '常见问题', link: '/zh-CN/faq' },
          { text: '许可证', link: '/zh-CN/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: zhSidebar,
        outline: { label: '本页目录', level: [2, 3] },
        docFooter: { prev: '上一页', next: '下一页' },
        lastUpdated: { text: '最后更新' },
        returnToTopLabel: '回到顶部',
        sidebarMenuLabel: '菜单',
        darkModeSwitchLabel: '外观',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: '在 GitHub 上编辑此页'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    },
    'zh-TW': {
      label: '繁體中文',
      lang: 'zh-TW',
      title: 'MySekaiMapper',
      description: '《世界計畫 繽紛舞台！feat. 初音未來》MySekai 採集點地圖產生與自動推播工具',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh-TW/guide/introduction', activeMatch: '/zh-TW/guide/' },
          { text: '常見問題', link: '/zh-TW/faq' },
          { text: '授權條款', link: '/zh-TW/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: zhTWSidebar,
        outline: { label: '本頁目錄', level: [2, 3] },
        docFooter: { prev: '上一頁', next: '下一頁' },
        lastUpdated: { text: '最後更新' },
        returnToTopLabel: '回到頂部',
        sidebarMenuLabel: '選單',
        darkModeSwitchLabel: '外觀',
        lightModeSwitchTitle: '切換到淺色模式',
        darkModeSwitchTitle: '切換到深色模式',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: '在 GitHub 上編輯此頁'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    },
    'ja-JP': {
      label: '日本語',
      lang: 'ja-JP',
      title: 'MySekaiMapper',
      description: '「プロジェクトセカイ カラフルステージ！」MySekai 採集ポイントマップ生成・自動プッシュツール',
      themeConfig: {
        nav: [
          { text: 'ガイド', link: '/ja-JP/guide/introduction', activeMatch: '/ja-JP/guide/' },
          { text: 'よくある質問', link: '/ja-JP/faq' },
          { text: 'ライセンス', link: '/ja-JP/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: jaJPSidebar,
        outline: { label: 'このページの目次', level: [2, 3] },
        docFooter: { prev: '前のページ', next: '次のページ' },
        lastUpdated: { text: '最終更新' },
        returnToTopLabel: 'トップへ戻る',
        sidebarMenuLabel: 'メニュー',
        darkModeSwitchLabel: 'テーマ',
        lightModeSwitchTitle: 'ライトモードに切り替え',
        darkModeSwitchTitle: 'ダークモードに切り替え',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: 'GitHub でこのページを編集'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    },
    'ko-KR': {
      label: '한국어',
      lang: 'ko-KR',
      title: 'MySekaiMapper',
      description: '프로젝트 세카이 MySekai 수집 포인트 지도 생성 및 자동 푸시 도구',
      themeConfig: {
        nav: [
          { text: '가이드', link: '/ko-KR/guide/introduction', activeMatch: '/ko-KR/guide/' },
          { text: '자주 묻는 질문', link: '/ko-KR/faq' },
          { text: '라이선스', link: '/ko-KR/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: koKRSidebar,
        outline: { label: '이 페이지 목차', level: [2, 3] },
        docFooter: { prev: '이전 페이지', next: '다음 페이지' },
        lastUpdated: { text: '마지막 업데이트' },
        returnToTopLabel: '맨 위로',
        sidebarMenuLabel: '메뉴',
        darkModeSwitchLabel: '테마',
        lightModeSwitchTitle: '라이트 모드로 전환',
        darkModeSwitchTitle: '다크 모드로 전환',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: 'GitHub에서 이 페이지 편집'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    }
  }
})
