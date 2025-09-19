import {computed, Directive, DOCUMENT, effect, ElementRef, inject, input, model, OnInit} from '@angular/core';
import {SettingsService} from "../_services/settings.service";

const DAY = 1000 * 60 * 60 * 24;

@Directive({
  selector: '[appUpdateBadge]'
})
export class UpdateBadgeDirective implements OnInit {

  private readonly settingsService = inject(SettingsService);
  private readonly document = inject(DOCUMENT);
  private readonly elementRef = inject(ElementRef);

  metadata = computed(() => this.settingsService.config()?.metadata)
  shouldDisplay = computed(() => {
    const metadata = this.metadata();
    const version = this.version();

    if (!metadata) return false;

    if (metadata.version !== version) return false;

    return new Date().getTime() - new Date(metadata.lastUpdateDate).getTime() < DAY;
  });

  id = model<string>();
  version = input.required<string>();
  badgePosition = input<'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'>('top-right');

  constructor() {
    effect(() => {
      if (!this.id()) {
        this.id.set(this.generateId());
      }
    });

    effect(() => {
      if (this.shouldDisplay()) {
        this.injectBadge();
      } else {
        this.removeBadge();
      }
    });
  }

  ngOnInit(): void {
    // Ensure parent element has relative positioning for badge positioning
    const parentElement = this.elementRef.nativeElement.parentElement;
    if (parentElement) {
      const computedStyle = window.getComputedStyle(parentElement);
      if (computedStyle.position === 'static') {
        parentElement.style.position = 'relative';
      }
    }
  }

  private injectBadge() {
    const id = this.id();
    if (!id) return;

    const element = this.document.getElementById(id);
    if (element) {
      return;
    }

    const parentElement = this.elementRef.nativeElement.parentElement;
    if (!parentElement) return;

    const badge = this.document.createElement('div');
    badge.id = id;
    badge.className = 'update-badge';
    badge.textContent = 'New!';

    this.applyBadgeStyles(badge);
    parentElement.appendChild(badge);

    requestAnimationFrame(() => {
      badge.style.opacity = '1';
      badge.style.transform = this.getTransformForPosition(this.badgePosition()) + ' scale(1)';
    });
  }

  private removeBadge() {
    const id = this.id();
    if (!id) return;

    const badge = this.document.getElementById(id);
    if (badge) {
      badge.style.opacity = '0';
      badge.style.transform = this.getTransformForPosition(this.badgePosition()) + ' scale(0.8)';

      setTimeout(() => {
        if (badge.parentNode) {
          badge.parentNode.removeChild(badge);
        } else {
          badge.remove();
        }
      }, 200);
    }
  }

  private applyBadgeStyles(badge: HTMLElement) {
    const position = this.badgePosition();

    Object.assign(badge.style, {
      position: 'absolute',
      zIndex: '1000',
      backgroundColor: 'var(--primary-color)',
      color: 'var(--text-light)',
      fontSize: '10px',
      fontWeight: '700',
      fontFamily: 'inherit',
      padding: '3px 6px',
      borderRadius: '10px',
      textAlign: 'center',
      lineHeight: '1',
      letterSpacing: '0.5px',
      textTransform: 'uppercase',
      cursor: 'default',
      userSelect: 'none',
      boxShadow: '0 2px 4px var(--shadow-medium)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      opacity: '0',
      transform: this.getTransformForPosition(position) + ' scale(0.8)',
      transition: 'all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1)',
      pointerEvents: 'none'
    });

    this.applyPositionStyles(badge, position);
  }

  private applyPositionStyles(badge: HTMLElement, position: string) {
    switch (position) {
      case 'top-right':
        Object.assign(badge.style, {
          top: '-8px',
          right: '-8px'
        });
        break;
      case 'top-left':
        Object.assign(badge.style, {
          top: '-8px',
          left: '-8px'
        });
        break;
      case 'bottom-right':
        Object.assign(badge.style, {
          bottom: '-8px',
          right: '-8px'
        });
        break;
      case 'bottom-left':
        Object.assign(badge.style, {
          bottom: '-8px',
          left: '-8px'
        });
        break;
    }
  }

  private getTransformForPosition(position: string): string {
    switch (position) {
      case 'top-right':
        return 'translate(50%, -50%)';
      case 'top-left':
        return 'translate(-50%, -50%)';
      case 'bottom-right':
        return 'translate(50%, 50%)';
      case 'bottom-left':
        return 'translate(-50%, 50%)';
      default:
        return 'translate(50%, -50%)';
    }
  }

  private generateId(): string {
    if (crypto && crypto.randomUUID) {
      return crypto.randomUUID();
    }

    // non secure connections
    return 'id-' + Math.random().toString(36).substr(2, 9) + '-' + Date.now().toString(36);
  }

}
