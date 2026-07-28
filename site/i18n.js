const translations = {
  en: {
    'document.title': 'Hometab - Your web, organized locally',
    'document.description': 'Hometab is a local-first personal start page delivered as a single executable.',
    'document.ogDescription': 'A local-first personal start page delivered as a single executable.',
    'nav.ariaLabel': 'Primary navigation',
    'nav.capabilities': 'Capabilities',
    'nav.delivery': 'Delivery',
    'language.ariaLabel': 'Language',
    'hero.eyebrow': 'Local-first personal start page',
    'hero.copy': 'Your links, search engines, saved requests, and visual workspace in one private application. One executable, one local database, no account required.',
    'hero.download': 'Download',
    'hero.source': 'View source',
    'capabilities.eyebrow': 'Built for everyday browsing',
    'capabilities.title': 'A calmer home for the web.',
    'capabilities.copy': 'Fast to open, easy to move, and entirely yours.',
    'feature.organize.title': 'Organize naturally',
    'feature.organize.copy': 'Group links, create reusable flows, and reorder everything around the way you work.',
    'feature.search.title': 'Search your way',
    'feature.search.copy': 'Switch between built-in and custom search engines without leaving your starting point.',
    'feature.local.title': 'Keep data local',
    'feature.local.copy': 'Store everything in a local SQLite database with versioned export, import, and restore tools.',
    'product.ariaLabel': 'Hometab application preview',
    'product.eyebrow': 'The actual application',
    'product.title': 'One interface from focused search to full dashboard.',
    'product.imageAlt': 'Hometab dashboard showing grouped links and search',
    'delivery.eyebrow': 'Delivery',
    'delivery.title': 'Start with one file.',
    'delivery.copy': 'Run Hometab directly. It selects an available local port and opens the right page automatically.',
    'delivery.binary.title': 'Single executable',
    'delivery.binary.copy': 'macOS, Linux, and Windows builds',
    'delivery.available': 'Available',
    'delivery.macos.title': 'macOS menu bar',
    'delivery.macos.copy': 'Native status and lifecycle controls',
    'delivery.planned': 'Planned',
    'delivery.windows.title': 'Windows desktop app',
    'delivery.windows.copy': 'Native background and window experience',
    'footer.copy': 'Local-first. Open source. MIT licensed.',
  },
  'zh-CN': {
    'document.title': 'Hometab - 本地优先的个人起始页',
    'document.description': 'Hometab 是以单一可执行文件交付的本地优先个人起始页。',
    'document.ogDescription': '以单一可执行文件交付的本地优先个人起始页。',
    'nav.ariaLabel': '主导航',
    'nav.capabilities': '核心能力',
    'nav.delivery': '交付方式',
    'language.ariaLabel': '语言',
    'hero.eyebrow': '本地优先的个人起始页',
    'hero.copy': '将链接、搜索引擎、固定请求与视觉工作区集中在一个私密应用中。一个可执行文件，一个本地数据库，无需账号。',
    'hero.download': '下载',
    'hero.source': '查看源码',
    'capabilities.eyebrow': '为日常浏览而生',
    'capabilities.title': '更从容的网络入口。',
    'capabilities.copy': '快速打开，灵活整理，完全属于你。',
    'feature.organize.title': '自然整理',
    'feature.organize.copy': '按你的工作方式组织链接、创建可复用流程，并自由调整每一项内容的顺序。',
    'feature.search.title': '按你的方式搜索',
    'feature.search.copy': '在内置和自定义搜索引擎之间快速切换，无需离开你的起始页。',
    'feature.local.title': '数据留在本地',
    'feature.local.copy': '所有内容存储在本地 SQLite 数据库，并支持版本化导出、导入与恢复。',
    'product.ariaLabel': 'Hometab 应用界面预览',
    'product.eyebrow': '真实应用界面',
    'product.title': '从专注搜索到完整仪表盘，一套界面即可切换。',
    'product.imageAlt': '展示分组链接与搜索功能的 Hometab 仪表盘',
    'delivery.eyebrow': '交付方式',
    'delivery.title': '从一个文件开始。',
    'delivery.copy': '直接运行 Hometab。它会自动选择可用的本地端口，并打开正确的页面。',
    'delivery.binary.title': '单一可执行文件',
    'delivery.binary.copy': '提供 macOS、Linux 与 Windows 构建',
    'delivery.available': '已可用',
    'delivery.macos.title': 'macOS 状态栏应用',
    'delivery.macos.copy': '原生状态展示与生命周期控制',
    'delivery.planned': '规划中',
    'delivery.windows.title': 'Windows 桌面应用',
    'delivery.windows.copy': '原生后台运行与窗口体验',
    'footer.copy': '本地优先。开源。MIT 许可。',
  },
}

const storageKey = 'hometab-site-language'
const supportedLanguages = Object.keys(translations)

function normalizeLanguage(value) {
  if (!value) return null
  const normalized = value.toLowerCase()
  if (normalized.startsWith('zh')) return 'zh-CN'
  if (normalized.startsWith('en')) return 'en'
  return null
}

function storedLanguage() {
  try {
    return normalizeLanguage(localStorage.getItem(storageKey))
  } catch {
    return null
  }
}

function initialLanguage() {
  const queryLanguage = normalizeLanguage(new URLSearchParams(window.location.search).get('lang'))
  return queryLanguage || storedLanguage() || normalizeLanguage(navigator.language) || 'en'
}

function setMetadata(dictionary) {
  document.title = dictionary['document.title']
  document.querySelector('meta[name="description"]')?.setAttribute('content', dictionary['document.description'])
  document.querySelector('meta[property="og:description"]')?.setAttribute('content', dictionary['document.ogDescription'])
}

function setLanguage(language, updateUrl = false) {
  const selected = supportedLanguages.includes(language) ? language : 'en'
  const dictionary = translations[selected]

  document.documentElement.lang = selected
  setMetadata(dictionary)

  document.querySelectorAll('[data-i18n]').forEach((element) => {
    const value = dictionary[element.dataset.i18n]
    if (value) element.textContent = value
  })

  document.querySelectorAll('[data-i18n-aria-label]').forEach((element) => {
    const value = dictionary[element.dataset.i18nAriaLabel]
    if (value) element.setAttribute('aria-label', value)
  })

  document.querySelectorAll('[data-i18n-alt]').forEach((element) => {
    const value = dictionary[element.dataset.i18nAlt]
    if (value) element.setAttribute('alt', value)
  })

  document.querySelectorAll('[data-language]').forEach((button) => {
    button.setAttribute('aria-pressed', String(button.dataset.language === selected))
  })

  try {
    localStorage.setItem(storageKey, selected)
  } catch {
    // The language still applies when storage is unavailable.
  }

  if (updateUrl) {
    const url = new URL(window.location.href)
    url.searchParams.set('lang', selected)
    history.replaceState(null, '', url)
  }
}

document.querySelectorAll('[data-language]').forEach((button) => {
  button.addEventListener('click', () => setLanguage(button.dataset.language, true))
})

setLanguage(initialLanguage())
