import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from './ui/dialog';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import type { LinkGroup } from '../types/links';
import { useTranslation } from 'react-i18next';

interface CreateGroupDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onCreateGroup: (group: Omit<LinkGroup, 'id' | 'created_at' | 'updated_at'>) => void;
  editingGroup?: LinkGroup | null;
}

export function CreateGroupDialog({
  isOpen,
  onOpenChange,
  onCreateGroup,
  editingGroup
}: CreateGroupDialogProps) {
  const { t } = useTranslation();
  const [name, setName] = useState(editingGroup?.name || '');
  const [description, setDescription] = useState(editingGroup?.description || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    onCreateGroup({
      name: name.trim(),
      description: description.trim() || undefined,
      order_index: editingGroup?.order_index || 0,
    });

    setName('');
    setDescription('');
    onOpenChange(false);
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setName(editingGroup?.name || '');
      setDescription(editingGroup?.description || '');
    }
    onOpenChange(open);
  };

  React.useEffect(() => {
    if (editingGroup) {
      setName(editingGroup.name);
      setDescription(editingGroup.description || '');
    } else {
      setName('');
      setDescription('');
    }
  }, [editingGroup]);

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {editingGroup ? t('group.editTitle') : t('group.createTitle')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="group-name">{t('group.nameLabel')}</Label>
            <Input
              id="group-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('group.namePlaceholder')}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="group-description">{t('group.descriptionLabel')}</Label>
            <Input
              id="group-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('group.descriptionPlaceholder')}
            />
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
            >
              {t('group.cancel')}
            </Button>
            <Button type="submit">
              {editingGroup ? t('group.save') : t('group.create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
