import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ProdutoService } from '../../services/produto.service';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { Produto } from '../../models/produto.model';
import { NotaFiscal } from '../../models/nota-fiscal.model';

type Aba = 'produtos' | 'notas';

@Component({
  selector: 'app-impressao',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './impressao.component.html',
  styleUrl: './impressao.component.css',
})
export class ImpressaoComponent implements OnInit {
  abaAtiva: Aba = 'produtos';
  produtos: Produto[] = [];
  notas: NotaFiscal[] = [];
  carregando = false;

  constructor(
    private produtoService: ProdutoService,
    private notaFiscalService: NotaFiscalService,
  ) {}

  ngOnInit(): void {
    this.carregando = true;
    this.produtoService.listar().subscribe({
      next: (produtos) => (this.produtos = produtos),
      complete: () => (this.carregando = false),
    });
    this.notaFiscalService.listar().subscribe({
      next: (notas) => (this.notas = notas),
    });
  }

  selecionarAba(aba: Aba): void {
    this.abaAtiva = aba;
  }

  imprimir(): void {
    window.print();
  }

  get dataEmissao(): Date {
    return new Date();
  }

  itensDaNota(nota: NotaFiscal): string {
    if (!nota.itens?.length) return '—';
    return nota.itens.map((item) => `${item.descricao} x${item.quantidade}`).join(', ');
  }
}
