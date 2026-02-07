import os

# Create a simple SVG icon and convert it
svg_content = '''<?xml version="1.0" encoding="UTF-8"?>
<svg width="1024" height="1024" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#0088cc;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#00a8ff;stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="1024" height="1024" rx="230" fill="url(#grad)"/>
  <path d="M512 256 L768 512 L512 768 L512 600 L256 600 L256 424 L512 424 Z" fill="white" transform="translate(0, 50)"/>
</svg>'''

with open('appicon.svg', 'w') as f:
    f.write(svg_content)

print("Icon SVG created")
