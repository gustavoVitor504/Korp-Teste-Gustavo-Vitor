import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ProdutoService } from '../../services/produto.service';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { Produto } from '../../models/produto.model';
import { NotaFiscal } from '../../models/nota-fiscal.model';

type Aba = 'produtos' | 'notas';   // aba so pode assumir 2 valores

@Component({
  selector: 'app-impressao',
  standalone: true,
  imports: [CommonModule],  // declarado para usar o Date e if e for
  templateUrl: './impressao.component.html',
  styleUrl: './impressao.component.css',
})
export class ImpressaoComponent implements OnInit {
  abaAtiva: Aba = 'produtos';      // aba inicia com produtos
  produtos: Produto[] = [];
  notas: NotaFiscal[] = [];
  carregando = false; 

  constructor(         // injeção de dependências para componentes não criarem dependências
    private produtoService: ProdutoService,
    private notaFiscalService: NotaFiscalService,
  ) {}

  ngOnInit(): void {
    this.carregando = true;  // ao iniciar carregando fica true e começar a buscar os elementos
    this.produtoService.listar().subscribe({ // primeiro elemento é chamado a service de produto e usado subscribe para escrever na página
      next: (produtos) => (this.produtos = produtos),
      complete: () => (this.carregando = false), // ao completar atualiza carregando para false
    }); // somente produto controla o carregando por enquanto
    this.notaFiscalService.listar().subscribe({ // chama service da nota fiscal para escrever as notas 
      next: (notas) => (this.notas = notas),
    });
  }

  selecionarAba(aba: Aba): void {
    this.abaAtiva = aba;   // abaAtiva recebe a aba que foi clicada
  }

  imprimir(): void {
    window.print();   // função de imprimir a tela
  }

  get dataEmissao(): Date {
    return new Date();    // obtem data e hora atual
  }

  itensDaNota(nota: NotaFiscal): string {
    if (!nota.itens?.length) return '—';
    return nota.itens.map((item) => `${item.descricao} x${item.quantidade}`).join(', ');
  }  // map transforma os itens em string para apresentar na tela
}
