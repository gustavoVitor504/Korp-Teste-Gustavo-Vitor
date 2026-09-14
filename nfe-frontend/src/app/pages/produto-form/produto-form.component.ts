import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ProdutoService } from '../../services/produto.service';
import { Produto } from '../../models/produto.model';

@Component({
  selector: 'app-produto-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './produto-form.component.html',
  styleUrl: './produto-form.component.css',
})
export class ProdutoFormComponent implements OnInit {
  produtos: Produto[] = [];
  carregando = false;
  salvando = false;
  mensagem: { tipo: 'ok' | 'erro'; texto: string } | null = null;

  private fb = inject(FormBuilder); // injeção de dependências
  private produtoService = inject(ProdutoService);

  form = this.fb.group({ // construtor do formulario 
    codigo: ['', [Validators.required, Validators.minLength(1)]],
    descricao: ['', [Validators.required, Validators.minLength(2)]],
    saldo: [0, [Validators.required, Validators.min(0)]],
  });

  ngOnInit(): void { // inicializa o componente e carrega os produtos
    this.carregarProdutos();
  }

  carregarProdutos(): void {
    this.carregando = true; // declara que está carregando
    this.produtoService.listar().subscribe({ // chama service para escrever os produtos
      next: (produtos) => {
        this.produtos = produtos; 
        this.carregando = false; // recebe os produtos e declara que terminou carregamento
      },
      error: () => { // tratamento de erro
        this.carregando = false;
        this.mensagem = { tipo: 'erro', texto: 'Não foi possível carregar os produtos.' };
      },
    });
  }

  salvar(): void {
    if (this.form.invalid) { // se formulário estiver inválido confrome o builder da erro
      this.form.markAllAsTouched();
      return;
    }

    const produto: Produto = {  // recebe os valores do formulario
      codigo: this.form.value.codigo!.trim(),
      descricao: this.form.value.descricao!.trim(),
      saldo: this.form.value.saldo!,
    };

    this.salvando = true; // declara que está salvando
    this.mensagem = null; // limpa mensagem

    this.produtoService.criar(produto).subscribe({ // chama service para criar
      next: () => {
        this.salvando = false; // não recebe nada, apenas declara que terminou de salvar
        this.mensagem = { tipo: 'ok', texto: `Produto "${produto.descricao}" cadastrado.` };
        this.form.reset({ codigo: '', descricao: '', saldo: 0 }); // envia mensagem de sucesso e reseta o formulario
        this.carregarProdutos(); // carrega os produtos de novo
      },
      error: (err) => { // tratamento de erro
        this.salvando = false;
        const texto = err?.error?.message ?? 'Erro ao cadastrar produto. Verifique os dados.';
        this.mensagem = { tipo: 'erro', texto };
      },
    });
  }
}
