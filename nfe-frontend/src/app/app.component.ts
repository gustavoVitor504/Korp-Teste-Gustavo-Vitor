import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  template: `
    <nav class="app-nav">
      <div class="app-nav-inner">
        <span class="brand">
          <span class="brand-mark">NF</span>
          Controle de Notas Fiscais
        </span>
        <div class="links">
          <a routerLink="/produtos" routerLinkActive="active">Produtos</a>
          <a routerLink="/notas-fiscais" routerLinkActive="active">Notas Fiscais</a>
          <a routerLink="/impressao" routerLinkActive="active">Impressão</a>
        </div>
      </div>
    </nav>
    <router-outlet></router-outlet>
  `,
  styles: [`
    .app-nav {
      background: var(--ink);
      border-bottom: 3px double var(--paper);
    }
    .app-nav-inner {
      max-width: 960px;
      margin: 0 auto;
      padding: 0 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 60px;
    }
    .brand {
      color: var(--paper);
      font-family: var(--font-display);
      font-size: 17px;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .brand-mark {
      font-family: var(--font-mono);
      font-size: 11px;
      border: 1.5px solid var(--paper);
      border-radius: 2px;
      padding: 2px 5px;
      transform: rotate(-3deg);
    }
    .links { display: flex; gap: 4px; }
    .links a {
      color: var(--paper);
      text-decoration: none;
      font-size: 13px;
      font-weight: 500;
      padding: 8px 14px;
      border-radius: 3px;
      opacity: 0.72;
      transition: opacity 0.15s ease, background 0.15s ease;
    }
    .links a:hover { opacity: 1; background: rgba(246,244,236,0.08); }
    .links a.active { opacity: 1; background: rgba(246,244,236,0.14); }
  `],
})
export class AppComponent {}
