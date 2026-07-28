import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Copy, Trash2 } from 'lucide-react'
import type { Link, LinkFlow, LinkFlowWithLinks, LinkGroup } from '@/types/link'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

export interface LinkFormData {
  name: string
  url: string
  group_id: string
  flow_id: string | null
}

export function LinkEditDialog({
  isOpen, onOpenChange, link, groups, flows, onSave, onDelete, onCopy, selectedGroup, selectedFlow,
}: {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  link?: Link | null
  groups: LinkGroup[]
  flows: LinkFlowWithLinks[]
  onSave: (linkData: LinkFormData) => void
  onDelete?: (link: Link) => void
  onCopy?: (link: Link) => void
  selectedGroup?: LinkGroup | null
  selectedFlow?: LinkFlow | null
}) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [groupId, setGroupId] = useState('')
  const [flowId, setFlowId] = useState('')

  useEffect(() => {
    queueMicrotask(() => {
      if (link) {
        setName(link.name); setUrl(link.url); setGroupId(link.group_id || ''); setFlowId(link.flow_id || '')
      } else {
        setName(''); setUrl('')
        if (selectedGroup) {
          setGroupId(selectedGroup.id)
          setFlowId(selectedFlow?.group_id === selectedGroup.id ? selectedFlow.id : '')
        } else {
          setGroupId(groups.find((group) => group.name === '未分组')?.id || '')
          setFlowId('')
        }
      }
    })
  }, [link, groups, selectedGroup, selectedFlow])

  const close = (open: boolean) => {
    if (!open) {
      setName(link?.name || ''); setUrl(link?.url || '')
      setGroupId(link?.group_id || selectedGroup?.id || groups.find((group) => group.name === '未分组')?.id || '')
      setFlowId(link?.flow_id || selectedFlow?.id || '')
    }
    onOpenChange(open)
  }

  return (
    <Dialog open={isOpen} onOpenChange={close}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader><DialogTitle>{link ? t('links.editLink') : t('links.addLink')}</DialogTitle></DialogHeader>
        <form onSubmit={(event) => { event.preventDefault(); if (name.trim() && url.trim()) onSave({ name: name.trim(), url: url.trim(), group_id: groupId, flow_id: flowId || null }) }} className="space-y-4">
          <div className="space-y-2"><Label htmlFor="link-url">{t('links.linkUrlLabel')}</Label><Input id="link-url" type="url" value={url} onChange={(event) => setUrl(event.target.value)} placeholder={t('links.linkUrlPlaceholder')} required /></div>
          <div className="space-y-2"><Label htmlFor="link-name">{t('links.linkNameLabel')}</Label><Input id="link-name" value={name} onChange={(event) => setName(event.target.value)} placeholder={t('links.linkNamePlaceholder')} required /></div>
          <div className="space-y-2">
            <Label htmlFor="link-group">{t('links.groupLabel')}</Label>
            <select id="link-group" value={groupId} onChange={(event) => setGroupId(event.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" required>
              {groups.map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="link-flow">{t('links.flowLabel')}</Label>
            <select id="link-flow" value={flowId} onChange={(event) => setFlowId(event.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
              <option value="">{t('links.none')}</option>
              {flows.filter(({ flow }) => flow.group_id === groupId).map(({ flow }) => <option key={flow.id} value={flow.id}>{flow.name}</option>)}
            </select>
          </div>
          <DialogFooter className="flex justify-between">
            <div className="flex gap-2">
              {link && onDelete && <Button type="button" variant="destructive" size="sm" onClick={() => { onDelete(link); close(false) }}><Trash2 className="h-4 w-4 mr-1" />{t('links.delete')}</Button>}
              {link && onCopy && <Button type="button" variant="outline" size="sm" onClick={() => { onCopy(link); close(false) }}><Copy className="h-4 w-4 mr-1" />{t('links.copy')}</Button>}
            </div>
            <div className="flex gap-2"><Button type="button" variant="outline" onClick={() => close(false)}>{t('links.cancel')}</Button><Button type="submit">{link ? t('links.save') : t('links.add')}</Button></div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
