import { useTheme } from './ThemeProvider'

interface LogoProps {
  className?: string
  height?: number
  width?: number
}

export function Logo({ className = "", height = 32, width = 120 }: LogoProps) {
  const { theme } = useTheme()
  
  // Choose the appropriate logo based on theme
  const logoSrc = theme === 'dark' 
    ? '/logo dewsign white.svg' 
    : '/logo dewsign black.svg'

  return (
    <img 
      src={logoSrc} 
      alt="Vexa" 
      className={className}
      height={height}
      width={width}
      style={{ height: `${height}px`, width: `${width}px` }}
    />
  )
}
