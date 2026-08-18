import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'produtos', pathMatch: 'full' },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./pages/produto-form/produto-form.component').then((m) => m.ProdutoFormComponent),
  },
  {
    path: 'notas-fiscais',
    loadComponent: () =>
      import('./pages/nota-fiscal-form/nota-fiscal-form.component').then(
        (m) => m.NotaFiscalFormComponent,
      ),
  },
  {
    path: 'impressao',
    loadComponent: () =>
      import('./pages/impressao/impressao.component').then((m) => m.ImpressaoComponent),
  },
  { path: '**', redirectTo: 'produtos' },
];
