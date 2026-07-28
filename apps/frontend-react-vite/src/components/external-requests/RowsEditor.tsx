import { Plus, Trash2 } from 'lucide-react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import type { KeyValueRow, ParserRow } from './request-form-utils'

type Row = KeyValueRow | ParserRow

export function RowsEditor<T extends Row>({
  rows,
  columns,
  onChange,
  addLabel,
  removeLabel,
}: {
  rows: T[]
  columns: Array<{ key: keyof Omit<T, 'id'>; placeholder: string }>
  onChange: (rows: T[]) => void
  addLabel: string
  removeLabel: string
}) {
  const update = (id: string, key: keyof Omit<T, 'id'>, value: string) => {
    onChange(rows.map((row) => row.id === id ? { ...row, [key]: value } : row))
  }
  const remove = (id: string) => {
    const next = rows.filter((row) => row.id !== id)
    onChange(next.length > 0 ? next : [{ id: crypto.randomUUID(), ...Object.fromEntries(columns.map(({ key }) => [key, ''])) } as T])
  }
  const add = () => {
    onChange([...rows, { id: crypto.randomUUID(), ...Object.fromEntries(columns.map(({ key }) => [key, ''])) } as T])
  }

  return (
    <div className="space-y-2">
      {rows.map((row) => (
        <div key={row.id} className="grid grid-cols-[1fr_1fr_32px] gap-2">
          {columns.map(({ key, placeholder }) => (
            <Input
              key={String(key)}
              value={String(row[key as keyof T] ?? '')}
              placeholder={placeholder}
              onChange={(event) => update(row.id, key, event.target.value)}
            />
          ))}
          <Button type="button" size="icon" variant="ghost" title={removeLabel} aria-label={removeLabel} onClick={() => remove(row.id)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <Button type="button" size="sm" variant="outline" title={addLabel} onClick={add}>
        <Plus className="h-4 w-4" />{addLabel}
      </Button>
    </div>
  )
}
