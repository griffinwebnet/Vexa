// Utility functions for domain information

export interface DomainInfo {
  domain: string
  realm: string
}

export function getDomainInfoFromStorage(): DomainInfo | null {
  try {
    const stored = localStorage.getItem('vexa-domain-info')
    if (stored) {
      return JSON.parse(stored)
    }
  } catch (error) {
    console.error('Failed to parse stored domain info:', error)
  }
  return null
}

export function clearDomainInfoFromStorage(): void {
  localStorage.removeItem('vexa-domain-info')
}
