//go:build darwin

#import <Cocoa/Cocoa.h>
#include "_cgo_export.h"

@interface SQTray : NSObject
@property(strong) NSStatusItem *item;
@end

@implementation SQTray
- (void)onItem:(NSMenuItem *)sender {
  goTrayItem((int)sender.tag);
}
@end

static SQTray *tray;

// Boo's silhouette as a template image, so macOS tints it for light and
// dark menu bars automatically.
static NSImage *booImage(void) {
  NSImage *img = [NSImage imageWithSize:NSMakeSize(18, 18)
                                flipped:YES
                         drawingHandler:^BOOL(NSRect rect) {
    NSBezierPath *body = [NSBezierPath bezierPath];
    [body moveToPoint:NSMakePoint(3, 16)];
    [body lineToPoint:NSMakePoint(3, 9)];
    [body appendBezierPathWithArcWithCenter:NSMakePoint(9, 9)
                                     radius:6
                                 startAngle:180
                                   endAngle:0
                                  clockwise:NO];
    [body lineToPoint:NSMakePoint(15, 16)];
    // Wavy hem.
    [body lineToPoint:NSMakePoint(13, 14.5)];
    [body lineToPoint:NSMakePoint(11, 16)];
    [body lineToPoint:NSMakePoint(9, 14.5)];
    [body lineToPoint:NSMakePoint(7, 16)];
    [body lineToPoint:NSMakePoint(5, 14.5)];
    [body closePath];
    [[NSColor blackColor] setFill];
    [body fill];

    // Eyes are punched out of the body.
    [[NSGraphicsContext currentContext] setCompositingOperation:NSCompositingOperationClear];
    [[NSBezierPath bezierPathWithOvalInRect:NSMakeRect(6, 7.5, 2, 3)] fill];
    [[NSBezierPath bezierPathWithOvalInRect:NSMakeRect(10, 7.5, 2, 3)] fill];
    return YES;
  }];
  img.template = YES;
  return img;
}

void trayStart(const char *tooltip, const char **labels, int count) {
  NSString *tip = [NSString stringWithUTF8String:tooltip];
  NSMutableArray<NSString *> *titles = [NSMutableArray arrayWithCapacity:count];
  for (int i = 0; i < count; i++) {
    [titles addObject:[NSString stringWithUTF8String:labels[i]]];
  }

  dispatch_async(dispatch_get_main_queue(), ^{
    // Wails makes the app a regular Dock app at launch; SideQuest lives in
    // the menu bar instead.
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

    if (tray != nil) {
      return;
    }
    tray = [SQTray new];
    tray.item = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];
    tray.item.button.image = booImage();
    tray.item.button.toolTip = tip;

    NSMenu *menu = [NSMenu new];
    menu.autoenablesItems = NO;
    for (NSUInteger i = 0; i < titles.count; i++) {
      if (titles[i].length == 0) {
        [menu addItem:[NSMenuItem separatorItem]];
        continue;
      }
      NSMenuItem *mi = [[NSMenuItem alloc] initWithTitle:titles[i]
                                                  action:@selector(onItem:)
                                           keyEquivalent:@""];
      mi.target = tray;
      mi.tag = (NSInteger)i;
      [menu addItem:mi];
    }
    tray.item.menu = menu;
  });
}

void traySetLabels(const char **labels, int count) {
  NSMutableArray<NSString *> *titles = [NSMutableArray arrayWithCapacity:count];
  for (int i = 0; i < count; i++) {
    [titles addObject:[NSString stringWithUTF8String:labels[i]]];
  }
  dispatch_async(dispatch_get_main_queue(), ^{
    if (tray == nil) {
      return;
    }
    for (NSMenuItem *mi in tray.item.menu.itemArray) {
      if (!mi.isSeparatorItem && mi.tag >= 0 && mi.tag < (NSInteger)titles.count) {
        mi.title = titles[mi.tag];
      }
    }
  });
}

void trayStop(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    if (tray != nil) {
      [[NSStatusBar systemStatusBar] removeStatusItem:tray.item];
      tray = nil;
    }
  });
}
