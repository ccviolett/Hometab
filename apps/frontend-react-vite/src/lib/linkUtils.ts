// 链接打开方式的工具函数

export type LinkOpenMode = 'new-tab' | 'current-tab' | 'new-window';

/**
 * 获取当前设置的链接打开方式
 */
export function getLinkOpenMode(): LinkOpenMode {
  // 强制设置为当前页面打开，覆盖之前的设置
  localStorage.setItem('link-open-mode', 'current-tab');
  return 'current-tab';
}

/**
 * 根据设置打开链接
 */
export function openLink(url: string, mode?: LinkOpenMode) {
  const linkMode = mode || getLinkOpenMode();

  switch (linkMode) {
    case 'current-tab':
      window.location.href = url;
      break;
    case 'new-window':
      window.open(url, '_blank', 'noopener,noreferrer');
      break;
    case 'new-tab':
    default:
      window.open(url, '_blank', 'noopener,noreferrer');
      break;
  }
}

/**
 * 获取链接的target和rel属性
 */
export function getLinkAttributes(mode?: LinkOpenMode): { target?: string; rel?: string } {
  const linkMode = mode || getLinkOpenMode();

  switch (linkMode) {
    case 'current-tab':
      return {};
    case 'new-window':
    case 'new-tab':
    default:
      return { target: '_blank', rel: 'noopener noreferrer' };
  }
}
