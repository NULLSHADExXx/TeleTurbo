import Cocoa

let size = CGSize(width: 1024, height: 1024)
let image = NSImage(size: size)

image.lockFocus()

// Background
let context = NSGraphicsContext.current!.cgContext
context.saveGState()

// Rounded rect path
let rect = CGRect(origin: .zero, size: size)
let path = NSBezierPath(roundedRect: rect, xRadius: 230, yRadius: 230)
path.addClip()

// Gradient
let gradient = NSGradient(colors: [
    NSColor(red: 0, green: 0.53, blue: 0.8, alpha: 1),
    NSColor(red: 0, green: 0.66, blue: 1, alpha: 1)
])
gradient?.draw(in: rect, angle: -45)

// Download arrow (white)
NSColor.white.setFill()
let arrowPath = NSBezierPath()
let center = CGPoint(x: 512, y: 512)
let arrowSize: CGFloat = 300

// Arrow body
arrowPath.move(to: CGPoint(x: center.x - arrowSize/4, y: center.y + arrowSize/3))
arrowPath.line(to: CGPoint(x: center.x + arrowSize/4, y: center.y + arrowSize/3))
arrowPath.line(to: CGPoint(x: center.x + arrowSize/4, y: center.y - arrowSize/6))
arrowPath.line(to: CGPoint(x: center.x + arrowSize/3, y: center.y - arrowSize/6))
arrowPath.line(to: CGPoint(x: center.x, y: center.y - arrowSize/2))
arrowPath.line(to: CGPoint(x: center.x - arrowSize/3, y: center.y - arrowSize/6))
arrowPath.line(to: CGPoint(x: center.x - arrowSize/4, y: center.y - arrowSize/6))
arrowPath.close()
arrowPath.fill()

// Arrow line
let lineRect = CGRect(x: center.x - arrowSize/12, y: center.y - arrowSize/2, width: arrowSize/6, height: arrowSize/3)
NSBezierPath(rect: lineRect).fill()

image.unlockFocus()

// Save as PNG
if let tiffData = image.tiffRepresentation,
   let bitmap = NSBitmapImageRep(data: tiffData),
   let pngData = bitmap.representation(using: .png, properties: [:]) {
    try? pngData.write(to: URL(fileURLWithPath: "appicon.png"))
    print("Icon saved to appicon.png")
}
