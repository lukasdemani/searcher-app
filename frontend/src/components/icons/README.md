# Icons Folder

Esta pasta contém todos os ícones SVG da aplicação organizados como componentes React reutilizáveis.

## Estrutura

Todos os ícones inline que estavam espalhados pela aplicação foram convertidos em componentes React individuais para melhor organização e reutilização.

### Ícones Comuns
- `CheckCircleIcon` - Ícone de check/confirmação
- `ArrowPathIcon` - Ícone de refresh/recarregar (com animação spin)
- `ClockIcon` - Ícone de relógio/pendente
- `XCircleIcon` - Ícone de erro/fechamento
- `ChevronDownIcon` - Seta para baixo (dropdown)
- `ChevronLeftIcon` - Seta para esquerda (voltar)
- `SpinnerIcon` - Ícone de carregamento com animação
- `ExclamationTriangleIcon` - Ícone de aviso/alerta
- `PlusIcon` - Ícone de adição/criar
- `LinkIcon` - Ícone de link/URL
- `TrashIcon` - Ícone de lixeira/deletar
- `PlayIcon` - Ícone de play/executar
- `DownloadIcon` - Ícone de download/exportar
- `UploadIcon` - Ícone de upload/importar

### Ícones do Dashboard
- `SortIcon` - Ícone de ordenação (neutro)
- `SortAscIcon` - Ícone de ordenação ascendente
- `SortDescIcon` - Ícone de ordenação descendente
- `StatsIcon` - Ícone de estatísticas
- `ProcessingIcon` - Ícone de processamento
- `SuccessIcon` - Ícone de sucesso
- `ErrorIcon` - Ícone de erro
- `NavigationIcon` - Ícone de navegação/logo
- `EmptyStateIcon` - Ícone de estado vazio
- `NoResultsIcon` - Ícone de sem resultados

### Ícones de Paginação
- `PreviousIcon` - Ícone de página anterior
- `NextIcon` - Ícone de próxima página

## Uso

### Importação Individual
```tsx
import { CheckCircleIcon } from '../icons';

<CheckCircleIcon className="h-4 w-4 text-green-600" />
```

### Importação Múltipla
```tsx
import { CheckCircleIcon, ErrorIcon, SpinnerIcon } from '../icons';
```

### Propriedades
Todos os ícones aceitam uma propriedade `className` opcional para personalização:

```tsx
interface IconProps {
  className?: string;
}
```

### Padrões de Tamanho
- `h-4 w-4` - Ícones pequenos (16px)
- `h-5 w-5` - Ícones médios (20px)
- `h-6 w-6` - Ícones grandes (24px)
- `h-12 w-12` - Ícones extra grandes (48px)

### Animações
Alguns ícones possuem animações CSS:
- `SpinnerIcon` - Sempre com animação de rotação
- `ArrowPathIcon` - Usado com classe `animate-spin` quando necessário

## Benefícios da Organização

1. **Reutilização** - Ícones podem ser facilmente reutilizados em toda a aplicação
2. **Consistência** - Todos os ícones seguem o mesmo padrão de implementação
3. **Manutenibilidade** - Mudanças em ícones podem ser feitas em um local central
4. **Performance** - Ícones são componentes otimizados do React
5. **Flexibilidade** - Fácil customização através de propriedades CSS

## Adicionando Novos Ícones

1. Crie um novo arquivo na pasta `icons/` seguindo o padrão `NomeIcon.tsx`
2. Implemente o componente seguindo a estrutura padrão
3. Adicione a exportação no arquivo `index.ts`
4. Use o ícone importando da pasta `icons`

### Exemplo de Novo Ícone

```tsx
import React from 'react';

interface NovoIconProps {
  className?: string;
}

const NovoIcon: React.FC<NovoIconProps> = ({ className = "h-4 w-4" }) => {
  return (
    <svg 
      className={className} 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
    >
      <path 
        strokeLinecap="round" 
        strokeLinejoin="round" 
        strokeWidth={2} 
        d="..." 
      />
    </svg>
  );
};

export default NovoIcon;
``` 