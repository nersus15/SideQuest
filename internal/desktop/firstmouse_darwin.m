//go:build darwin

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

static BOOL yes(id self, SEL _cmd, NSEvent *event) { return YES; }

void sqAcceptFirstMouse(void) {
  // Wails' WKWebView subclass; add or replace acceptsFirstMouse: on it.
  Class cls = NSClassFromString(@"WailsWebView");
  if (cls == Nil) {
    return;
  }
  // Copy the type encoding from NSView: BOOL differs between architectures.
  Method base = class_getInstanceMethod([NSView class], @selector(acceptsFirstMouse:));
  class_replaceMethod(cls, @selector(acceptsFirstMouse:), (IMP)yes, method_getTypeEncoding(base));
}
