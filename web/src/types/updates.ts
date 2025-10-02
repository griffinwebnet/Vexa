export interface SystemVersion {
  component: string
  version: string
}

export interface GitHubRelease {
  tag_name: string
  name: string
  body: string
  published_at: string
  html_url: string
  assets: Array<{
    name: string
    browser_download_url: string
    size: number
  }>
}

export interface UpdateInfo {
  versions: SystemVersion[]
  latest_version: string
  update_available: boolean
  status: 'Development Version' | 'Update Available' | 'Up to Date'
  latest_release: GitHubRelease
  error?: string
}
