import { Button } from '../ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '../ui/dialog';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Plus, Edit2, Trash2, ExternalLink, FolderPlus, Folder, AlertTriangle, ArrowUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, ChevronDown, ChevronUp, Play, X } from 'lucide-react';
import { CreateGroupDialog } from '../CreateGroupDialog';
import { openLink } from '@/lib/linkUtils';
import { FaviconIcon } from './FaviconIcon';
import { LinkEditDialog } from './LinkEditDialog';
import { useLinksSection } from './useLinksSection';
import { SortableHandle } from './SortableHandle';
import { DndContext, KeyboardSensor, PointerSensor, TouchSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core';
import { SortableContext, rectSortingStrategy, sortableKeyboardCoordinates, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { moveId } from './reorder';

function withVersion(url: string, version: number) {
  if (version <= 0) return url;
  return url + (url.includes('?') ? '&' : '?') + `v=${version}`;
}

export function LinksSection() {
  const {
    t, groupedLinks, error, isGroupDialogOpen, setIsGroupDialogOpen, editingGroup,
    isLinkDialogOpen, setIsLinkDialogOpen, editingLink, selectedFlowForNewLink,
    isFlowDialogOpen, setIsFlowDialogOpen, flowDialogGroupId, setFlowDialogGroupId, flowNameInput,
    setFlowNameInput, isFlowDetailOpen, setIsFlowDetailOpen, activeFlow, flowLinksToKeep,
    setFlowLinksToKeep, isFlowDeleteDialogOpen, setIsFlowDeleteDialogOpen, isDeleteGroupDialogOpen,
    setIsDeleteGroupDialogOpen, groupToDelete, deleteOption, setDeleteOption, selectedGroupForNewLink,
    sortingGroupId, isDeleteLinkDialogOpen, setIsDeleteLinkDialogOpen, linkToDelete, setLinkToDelete,
    collapsedGroups, sortingMode, setSortingMode, iconConflict, setIconConflict, iconVersion, isLoading,
    handleCreateGroup, handleEditGroup, handleDeleteGroup, handleConfirmDeleteGroup, handleEditLink,
    handleDeleteLink, handleCopyLink, handleConfirmDeleteLink, handleAddLink, handleAddLinkToGroup,
    handleAddLinkToFlow, handleFlowCreated, handleOpenFlowDetail, handleExecuteFlow, moveFlowLink,
    moveFlowLinkToStart, moveFlowLinkToEnd, handleRemoveLinkFromFlow, handleConfirmDeleteFlow,
    handleChooseIcon, handleToggleSorting, handleMoveLink, handleMoveLinkToStart, handleMoveLinkToEnd,
    handleSaveLink, toggleGroupCollapse, swapGroupOrder, moveGroupToStart, moveGroupToEnd,
    reorderGroups, reorderGroupLinks, reorderFlowLinks,
  } = useLinksSection()
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 180, tolerance: 6 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
  const expandedGroups = groupedLinks.filter(group => !collapsedGroups.has(group.group.id));
  const collapsedGroupsList = groupedLinks.filter(group => collapsedGroups.has(group.group.id));

  const reorderedIds = (ids: string[], event: DragEndEvent) => {
    if (!event.over || event.active.id === event.over.id) return null;
    return moveId(ids, String(event.active.id), String(event.over.id));
  };

  if (isLoading) return <div>Loading...</div>;

  return (
    <div className="space-y-4">
      {error && (
        <div className="bg-orange-100/80 border border-orange-300/50 text-orange-800 px-4 py-3 rounded-lg backdrop-blur-sm">
          <div className="flex items-center gap-2"><ExternalLink className="h-4 w-4 flex-shrink-0" /><span className="text-sm">{error}</span></div>
        </div>
      )}

      <div className="flex gap-3 flex-wrap items-center">
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => setIsGroupDialogOpen(true)} className="bg-white/20 hover:bg-white/30 border-white/30 text-white backdrop-blur-sm transition-colors cursor-pointer" variant="outline">
            <FolderPlus className="h-4 w-4 mr-2" />{t('links.createGroup')}
          </Button>
          <Button onClick={() => { setFlowDialogGroupId(''); setFlowNameInput(''); setIsFlowDialogOpen(true); }} className="bg-white/20 hover:bg-white/30 border-white/30 text-white backdrop-blur-sm transition-colors cursor-pointer" variant="outline">
            <Play className="h-4 w-4 mr-2" />{t('links.createFlow')}
          </Button>
          <Button onClick={handleAddLink} className="bg-white/20 hover:bg-white/30 border-white/30 text-white backdrop-blur-sm transition-colors cursor-pointer" variant="outline">
            <Plus className="h-4 w-4 mr-2" />{t('links.addLink')}
          </Button>
          <Button onClick={() => setSortingMode(!sortingMode)} className={`${sortingMode ? 'bg-white/40 border-white/50' : 'bg-white/20 hover:bg-white/30'} border-white/30 text-white backdrop-blur-sm transition-colors cursor-pointer`} variant="outline">
            <ArrowUpDown className="h-4 w-4 mr-2" />{sortingMode ? t('links.finishSorting') : t('links.sortGroups')}
          </Button>
        </div>

        {collapsedGroupsList.length > 0 && (
          <div className="flex gap-2 items-center">
            <div className="w-px h-6 bg-white/30"></div>
            {collapsedGroupsList.map((groupData) => (
              <Button key={groupData.group.id} variant="outline" size="sm" className="h-8 px-3 bg-white/20 border-white/30 text-white/90 hover:bg-white/30 hover:text-white cursor-pointer" onClick={() => toggleGroupCollapse(groupData.group.id)}>
                <Folder className="h-3 w-3 mr-1.5" />
                <span className="text-xs font-medium">{groupData.group.name}</span>
                {groupData.group.description && (<span className="ml-1.5 text-xs text-white/50 italic">· {groupData.group.description}</span>)}
                <span className="ml-1.5 text-xs text-white/60">({groupData.links.length})</span>
                <ChevronDown className="h-3 w-3 ml-1" />
              </Button>
            ))}
          </div>
        )}
      </div>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={(event) => {
          if (!sortingMode) return;
          const next = reorderedIds(expandedGroups.map((group) => group.group.id), event);
          if (next) void reorderGroups(next);
        }}
      >
        <SortableContext items={expandedGroups.map((group) => group.group.id)} strategy={rectSortingStrategy}>
      <div className="grid grid-cols-1 2xl:grid-cols-2 gap-6">
        {expandedGroups.map((groupData) => (
          <SortableHandle key={groupData.group.id} id={groupData.group.id} label={t('links.dragGroup')} className={`relative ${sortingMode ? '' : '[&>button:first-child]:hidden'}`}>
          <div className="bg-white/20 backdrop-blur-sm rounded-2xl p-4 border border-white/30 shadow-lg">
            <div className="flex items-center justify-between mb-3 pb-2 border-b border-white/20">
              <div className="flex items-center gap-2">
                <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/70 hover:text-white hover:bg-white/20 rounded-sm cursor-pointer" onClick={() => toggleGroupCollapse(groupData.group.id)}>
                  <ChevronUp className="h-3 w-3" />
                </Button>
                <div className="flex items-center gap-2">
                  <h3 className="text-base font-medium text-white/90 drop-shadow-sm">{groupData.group.name}</h3>
                  {groupData.group.description && (<span className="text-sm text-white/50 italic">· {groupData.group.description}</span>)}
                </div>
                <span className="text-xs text-white/60 bg-white/10 px-1.5 py-0.5 rounded-full">{groupData.links.length}</span>
                {sortingMode && (
                  <div className="flex ml-2 gap-0.5">
                    <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/60 hover:text-white/80 hover:bg-white/15 rounded-sm cursor-pointer" onClick={() => moveGroupToStart(groupData.group.id)} disabled={groupedLinks.findIndex(g => g.group.id === groupData.group.id) === 0} title={t('links.moveToFront')}><ChevronsLeft className="h-3 w-3" /></Button>
                    <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/45 hover:text-white/70 hover:bg-white/10 rounded-sm cursor-pointer" onClick={() => swapGroupOrder(groupData.group.id, 'up')} disabled={groupedLinks.findIndex(g => g.group.id === groupData.group.id) === 0} title={t('links.moveLeftOne')}><ChevronLeft className="h-3 w-3" /></Button>
                    <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/45 hover:text-white/70 hover:bg-white/10 rounded-sm cursor-pointer" onClick={() => swapGroupOrder(groupData.group.id, 'down')} disabled={groupedLinks.findIndex(g => g.group.id === groupData.group.id) === groupedLinks.length - 1} title={t('links.moveRightOne')}><ChevronRight className="h-3 w-3" /></Button>
                    <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/60 hover:text-white/80 hover:bg-white/15 rounded-sm cursor-pointer" onClick={() => moveGroupToEnd(groupData.group.id)} disabled={groupedLinks.findIndex(g => g.group.id === groupData.group.id) === groupedLinks.length - 1} title={t('links.moveToEnd')}><ChevronsRight className="h-3 w-3" /></Button>
                  </div>
                )}
              </div>
              <div className="flex gap-0.5 bg-white/30 rounded-md p-0.5">
                <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-green-600 hover:text-green-500 hover:bg-green-500/20 rounded-sm cursor-pointer" onClick={() => handleAddLinkToGroup(groupData.group)}><Plus className="h-3 w-3" /></Button>
                <Button size="sm" variant="ghost" className={`h-6 w-6 p-0 rounded-sm cursor-pointer ${sortingGroupId === groupData.group.id ? 'text-blue-600 bg-blue-500/20 hover:text-blue-500 hover:bg-blue-500/30' : 'text-blue-300 hover:text-blue-200 hover:bg-blue-500/20'}`} onClick={() => handleToggleSorting(groupData.group.id)}><ArrowUpDown className="h-3 w-3" /></Button>
                <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-white/70 hover:text-white hover:bg-white/20 rounded-sm cursor-pointer" onClick={() => handleEditGroup(groupData.group)}><Edit2 className="h-3 w-3" /></Button>
                <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-red-300 hover:text-red-200 hover:bg-red-500/20 rounded-sm cursor-pointer" onClick={() => handleDeleteGroup(groupData.group)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </div>

            {groupData.flows && groupData.flows.length > 0 && (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-3 gap-3 mb-4">
                {groupData.flows.map((flow) => (
                  <div key={flow.flow.id} className="group relative px-3 py-2.5 rounded-xl border border-white/30 bg-white/90 transition-all duration-200 overflow-hidden backdrop-blur-sm hover:bg-white hover:shadow-md cursor-pointer" onClick={() => handleExecuteFlow(flow)}>
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="text-sm font-medium text-gray-800 truncate flex items-center gap-1"><Play className="h-3 w-3 text-emerald-600" />{flow.flow.name}</h4>
                      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-gray-600 hover:text-gray-800 bg-white/90 hover:bg-white hover:shadow-md shadow-sm cursor-pointer transition-all duration-200" onClick={(e) => { e.stopPropagation(); handleOpenFlowDetail(flow); }} title={t('links.editFlow')}><Edit2 className="h-3 w-3" /></Button>
                      </div>
                    </div>
                    {flow.links.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {flow.links.slice(0, 6).map((link) => (<div key={link.id} className="w-5 h-5"><FaviconIcon url={link.url} size="small" version={iconVersion} /></div>))}
                        {flow.links.length > 6 && (<div className="w-5 h-5 flex items-center justify-center text-xs text-gray-500">+{flow.links.length - 6}</div>)}
                      </div>
                    ) : (<div className="text-xs text-gray-400">{t('links.noLinks')}</div>)}
                    <div className="absolute inset-0 bg-emerald-500/10 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                  </div>
                ))}
              </div>
            )}

            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={(event) => {
                if (sortingGroupId !== groupData.group.id) return;
                const ids = [...groupData.links].sort((a, b) => a.order_index - b.order_index).map((link) => link.id);
                const next = reorderedIds(ids, event);
                if (next) void reorderGroupLinks(groupData.group.id, next);
              }}
            >
              <SortableContext items={groupData.links.map((link) => link.id)} strategy={rectSortingStrategy}>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-3 gap-3">
              {[...groupData.links].sort((a, b) => a.order_index - b.order_index).map((link, index, sortedLinks) => {
                const isSortingMode = sortingGroupId === groupData.group.id;
                const canMoveLeft = index > 0;
                const canMoveRight = index < sortedLinks.length - 1;
                return (
                  <SortableHandle key={link.id} id={link.id} label={t('links.dragLink')} className={`relative ${isSortingMode ? '' : '[&>button:first-child]:hidden'}`}>
                  <div className={`group relative px-3 py-2.5 rounded-xl border border-white/30 bg-white/90 transition-all duration-200 overflow-hidden backdrop-blur-sm ${isSortingMode ? 'cursor-default pl-8' : 'hover:bg-white hover:shadow-md cursor-pointer'}`} onClick={isSortingMode ? undefined : () => openLink(link.url)}>
                    {isSortingMode ? (
                      <>
                        <div className="flex items-center gap-2"><FaviconIcon url={link.url} version={iconVersion} /><span className="text-sm text-gray-800 font-medium truncate">{link.name}</span></div>
                        <div className="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-white/95 via-white/90 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                        <div className="absolute right-2 top-1/2 transform -translate-y-1/2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                          <div className="flex gap-0.5 bg-gray-200/80 backdrop-blur-sm rounded p-0.5">
                            <Button size="sm" variant="ghost" className="h-5 w-5 p-0 text-gray-600 hover:text-gray-800 hover:bg-gray-300/50 rounded-sm cursor-pointer disabled:opacity-40" onClick={(e) => { e.stopPropagation(); if (canMoveLeft) handleMoveLinkToStart(link.id); }} disabled={!canMoveLeft} title={t('links.moveToFront')}><ChevronsLeft className="h-2.5 w-2.5" /></Button>
                            <Button size="sm" variant="ghost" className="h-5 w-5 p-0 text-gray-500 hover:text-gray-700 hover:bg-gray-300/30 rounded-sm cursor-pointer disabled:opacity-40" onClick={(e) => { e.stopPropagation(); if (canMoveLeft) handleMoveLink(link.id, 'left'); }} disabled={!canMoveLeft} title={t('links.moveLeftOne')}><ChevronLeft className="h-2.5 w-2.5" /></Button>
                            <Button size="sm" variant="ghost" className="h-5 w-5 p-0 text-gray-500 hover:text-gray-700 hover:bg-gray-300/30 rounded-sm cursor-pointer disabled:opacity-40" onClick={(e) => { e.stopPropagation(); if (canMoveRight) handleMoveLink(link.id, 'right'); }} disabled={!canMoveRight} title={t('links.moveRightOne')}><ChevronRight className="h-2.5 w-2.5" /></Button>
                            <Button size="sm" variant="ghost" className="h-5 w-5 p-0 text-gray-600 hover:text-gray-800 hover:bg-gray-300/50 rounded-sm cursor-pointer disabled:opacity-40" onClick={(e) => { e.stopPropagation(); if (canMoveRight) handleMoveLinkToEnd(link.id); }} disabled={!canMoveRight} title={t('links.moveToEnd')}><ChevronsRight className="h-2.5 w-2.5" /></Button>
                          </div>
                        </div>
                      </>
                    ) : (
                      <>
                        <div className="flex items-center gap-2"><FaviconIcon url={link.url} version={iconVersion} /><span className="text-sm text-gray-800 font-medium truncate">{link.name}</span></div>
                        <div className="absolute right-0 top-0 bottom-0 w-16 bg-gradient-to-l from-white via-white/95 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                        <div className="absolute right-2 top-1/2 transform -translate-y-1/2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                          <Button size="sm" variant="ghost" className="h-6 w-6 p-0 text-gray-600 hover:text-gray-700 hover:bg-gray-100 bg-white/90 hover:bg-white shadow-sm cursor-pointer transition-colors" onClick={(e) => { e.stopPropagation(); handleEditLink(link); }} title={t('links.editLink')}><Edit2 className="h-3 w-3" /></Button>
                        </div>
                      </>
                    )}
                  </div>
                  </SortableHandle>
                );
              })}
              {groupData.links.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center py-8 text-white/60">
                  <div className="w-12 h-12 rounded-full bg-white/10 flex items-center justify-center mb-3"><ExternalLink className="h-6 w-6" /></div>
                  <p className="text-sm">{t('links.noLinks')}</p>
                  <Button size="sm" variant="ghost" className="mt-2 text-white/70 hover:text-white hover:bg-white/20 cursor-pointer" onClick={handleAddLink}><Plus className="h-4 w-4 mr-1" />{t('links.addLink')}</Button>
                </div>
              )}
            </div>
              </SortableContext>
            </DndContext>
          </div>
          </SortableHandle>
        ))}
      </div>
        </SortableContext>
      </DndContext>

      <CreateGroupDialog isOpen={isGroupDialogOpen} onOpenChange={setIsGroupDialogOpen} onCreateGroup={handleCreateGroup} editingGroup={editingGroup} />
      <LinkEditDialog isOpen={isLinkDialogOpen} onOpenChange={setIsLinkDialogOpen} link={editingLink} groups={groupedLinks.map(g => g.group)} flows={groupedLinks.flatMap(g => g.flows)} onSave={handleSaveLink} onDelete={handleDeleteLink} onCopy={handleCopyLink} selectedGroup={selectedGroupForNewLink} selectedFlow={selectedFlowForNewLink} />

      {isDeleteGroupDialogOpen && groupToDelete && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center">
          <div className="fixed inset-0 bg-black/50" onClick={() => setIsDeleteGroupDialogOpen(false)} />
          <div className="relative bg-white rounded-lg p-6 w-full max-w-md mx-4 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="flex-shrink-0 w-10 h-10 bg-red-100 rounded-full flex items-center justify-center"><AlertTriangle className="h-5 w-5 text-red-600" /></div>
              <div><h3 className="text-lg font-semibold text-gray-900">{t('links.deleteGroup')}</h3><p className="text-sm text-gray-500">{t('links.deleteGroupConfirm', { name: groupToDelete.name })}</p></div>
            </div>
            {(() => { const groupData = groupedLinks.find(g => g.group.id === groupToDelete.id); const linkCount = groupData?.links.length || 0;
              return linkCount > 0 ? (
                <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-md">
                  <p className="text-sm text-amber-800 mb-3">{t('links.groupContainsLinks', { count: linkCount })}</p>
                  <div className="space-y-2">
                    <label className="flex items-center"><input type="radio" name="deleteOption" value="move" checked={deleteOption === 'move'} onChange={(e) => setDeleteOption(e.target.value as 'move' | 'delete')} className="mr-2" /><span className="text-sm text-gray-700">{t('links.moveToUngrouped')}</span></label>
                    <label className="flex items-center"><input type="radio" name="deleteOption" value="delete" checked={deleteOption === 'delete'} onChange={(e) => setDeleteOption(e.target.value as 'move' | 'delete')} className="mr-2" /><span className="text-sm text-red-700">{t('links.deleteAllLinks')}</span></label>
                  </div>
                </div>
              ) : (<p className="text-sm text-gray-600 mb-4">{t('links.groupEmptySafeDelete')}</p>);
            })()}
            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsDeleteGroupDialogOpen(false)}>{t('links.cancel')}</Button>
              <Button variant="destructive" onClick={handleConfirmDeleteGroup}>{t('links.confirmDelete')}</Button>
            </div>
          </div>
        </div>
      )}

      {isDeleteLinkDialogOpen && linkToDelete && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center">
          <div className="fixed inset-0 bg-black/50" onClick={() => setIsDeleteLinkDialogOpen(false)} />
          <div className="relative bg-white rounded-lg p-6 w-full max-w-md mx-4 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="flex-shrink-0 w-10 h-10 bg-red-100 rounded-full flex items-center justify-center"><AlertTriangle className="h-5 w-5 text-red-600" /></div>
              <div><h3 className="text-lg font-semibold text-gray-900">{t('links.deleteLink')}</h3><p className="text-sm text-gray-500">{t('links.deleteLinkConfirm')}</p></div>
            </div>
            <div className="mb-4 p-3 bg-gray-50 rounded-lg border">
              <div className="flex items-center gap-3"><FaviconIcon url={linkToDelete.url} version={iconVersion} /><div className="flex-1 min-w-0"><p className="text-sm font-medium text-gray-900 truncate">{linkToDelete.name}</p><p className="text-xs text-gray-500 truncate">{linkToDelete.url}</p></div></div>
            </div>
            <div className="text-sm text-gray-600 mb-4"><p>{t('links.linkDeleteWarning')}</p></div>
            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => { setIsDeleteLinkDialogOpen(false); setLinkToDelete(null); }}>{t('links.cancel')}</Button>
              <Button variant="destructive" onClick={handleConfirmDeleteLink}>{t('links.confirmDelete')}</Button>
            </div>
          </div>
        </div>
      )}

      {iconConflict && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center">
          <div className="fixed inset-0 bg-black/50" onClick={() => setIconConflict(null)} />
          <div className="relative bg-white rounded-lg p-6 w-full max-w-md mx-4 shadow-xl">
            <div className="mb-4">
              <h3 className="text-lg font-semibold text-gray-900">{t('links.chooseIcon')}</h3>
              <p className="text-sm text-gray-500 mt-1">{t('links.iconConflictDesc', { host: iconConflict.host })}</p>
            </div>
            <div className="grid grid-cols-2 gap-3 mb-5">
              <button type="button" className="rounded-lg border border-gray-200 p-4 hover:bg-gray-50 cursor-pointer text-left" onClick={() => handleChooseIcon('current')}>
                <div className="w-12 h-12 rounded-lg bg-gray-100 flex items-center justify-center mb-3 overflow-hidden">
                  <img src={withVersion(iconConflict.currentIconUrl, iconVersion)} alt={t('links.currentIcon')} className="w-8 h-8" />
                </div>
                <div className="text-sm font-medium text-gray-900">{t('links.keepCurrent')}</div>
              </button>
              <button type="button" className="rounded-lg border border-blue-200 p-4 hover:bg-blue-50 cursor-pointer text-left" onClick={() => handleChooseIcon('new')}>
                <div className="w-12 h-12 rounded-lg bg-blue-50 flex items-center justify-center mb-3 overflow-hidden">
                  <img src={withVersion(iconConflict.pendingIconUrl, iconVersion)} alt={t('links.newIcon')} className="w-8 h-8" />
                </div>
                <div className="text-sm font-medium text-gray-900">{t('links.useNewIcon')}</div>
              </button>
            </div>
            <div className="flex justify-end">
              <Button variant="outline" onClick={() => setIconConflict(null)}>{t('links.later')}</Button>
            </div>
          </div>
        </div>
      )}

      <Dialog open={isFlowDialogOpen} onOpenChange={setIsFlowDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{t('links.createFlow')}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="flow-group">{t('links.belongingGroup')}</Label>
              <select id="flow-group" value={flowDialogGroupId} onChange={(e) => setFlowDialogGroupId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" required>
                <option value="">{t('links.selectGroupPlaceholder')}</option>
                {groupedLinks.map((g) => (<option key={g.group.id} value={g.group.id}>{g.group.name}</option>))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="flow-name">{t('links.flowNameLabel')}</Label>
              <Input id="flow-name" value={flowNameInput} onChange={(e) => setFlowNameInput(e.target.value)} placeholder={t('links.flowNamePlaceholder')} />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setIsFlowDialogOpen(false)}>{t('links.cancel')}</Button>
            <Button onClick={handleFlowCreated} disabled={!flowNameInput.trim() || !flowDialogGroupId}>{t('links.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isFlowDetailOpen} onOpenChange={setIsFlowDetailOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader><DialogTitle>{activeFlow?.flow.name}</DialogTitle></DialogHeader>
          {activeFlow && (
            <div className="space-y-4 max-h-[60vh] overflow-y-auto">
              {activeFlow.links.length > 0 ? (
                <DndContext
                  sensors={sensors}
                  collisionDetection={closestCenter}
                  onDragEnd={(event) => {
                    const next = reorderedIds(activeFlow.links.map((link) => link.id), event);
                    if (next) void reorderFlowLinks(activeFlow.flow.id, next);
                  }}
                >
                  <SortableContext items={activeFlow.links.map((link) => link.id)} strategy={verticalListSortingStrategy}>
                <div className="space-y-3 pr-2">
                  {activeFlow.links.map((link, index, sortedLinks) => (
                    <SortableHandle key={link.id} id={link.id} label={t('links.dragLink')} className="relative">
                    <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3 pl-9 group hover:bg-gray-50 cursor-pointer transition-colors" onClick={() => openLink(link.url)}>
                      <div className="flex-shrink-0"><FaviconIcon url={link.url} version={iconVersion} /></div>
                      <div className="flex-1 min-w-0"><p className="text-sm font-medium text-gray-900 truncate">{link.name}</p><p className="text-xs text-gray-500 truncate">{link.url}</p></div>
                      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-blue-400 hover:text-blue-500 hover:bg-blue-50" disabled={index === 0} onClick={(e) => { e.stopPropagation(); moveFlowLinkToStart(activeFlow.flow.id, activeFlow.links, link.id); }} title={t('links.moveToFront')}><ChevronsLeft className="h-3.5 w-3.5" /></Button>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-blue-300 hover:text-blue-400 hover:bg-blue-50" disabled={index === 0} onClick={(e) => { e.stopPropagation(); moveFlowLink(activeFlow.flow.id, activeFlow.links, link.id, 'left'); }} title={t('links.moveLeftOne')}><ChevronLeft className="h-3.5 w-3.5" /></Button>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-blue-300 hover:text-blue-400 hover:bg-blue-50" disabled={index === sortedLinks.length - 1} onClick={(e) => { e.stopPropagation(); moveFlowLink(activeFlow.flow.id, activeFlow.links, link.id, 'right'); }} title={t('links.moveRightOne')}><ChevronRight className="h-3.5 w-3.5" /></Button>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-blue-400 hover:text-blue-500 hover:bg-blue-50" disabled={index === sortedLinks.length - 1} onClick={(e) => { e.stopPropagation(); moveFlowLinkToEnd(activeFlow.flow.id, activeFlow.links, link.id); }} title={t('links.moveToEnd')}><ChevronsRight className="h-3.5 w-3.5" /></Button>
                        <div className="w-px h-5 bg-gray-300 mx-0.5"></div>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={(e) => { e.stopPropagation(); handleEditLink(link); }} title={t('links.edit')}><Edit2 className="h-3.5 w-3.5" /></Button>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-red-600 hover:text-red-700" onClick={(e) => { e.stopPropagation(); handleRemoveLinkFromFlow(activeFlow.flow.id, link.id); }} title={t('links.remove')}><X className="h-3.5 w-3.5" /></Button>
                      </div>
                    </div>
                    </SortableHandle>
                  ))}
                </div>
                  </SortableContext>
                </DndContext>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-gray-500"><ExternalLink className="h-8 w-8 mb-3 text-gray-400" /><p className="text-sm">{t('links.noLinks')}</p></div>
              )}
              <div className="pt-4 mt-4 border-t border-gray-200 flex justify-between items-center">
                <Button variant="destructive" size="sm" onClick={() => setIsFlowDeleteDialogOpen(true)}><Trash2 className="h-4 w-4 mr-1" />{t('links.deleteFlow')}</Button>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => { if (activeFlow) { const group = groupedLinks.find(g => g.group.id === activeFlow.flow.group_id); if (group) { handleAddLinkToFlow(group.group, activeFlow.flow); setIsFlowDetailOpen(false); } } }}><Plus className="h-4 w-4 mr-1" />{t('links.addLink')}</Button>
                  <Button size="sm" className="bg-emerald-600 hover:bg-emerald-700 text-white" onClick={() => activeFlow && handleExecuteFlow(activeFlow)}><Play className="h-4 w-4 mr-1" />{t('links.executeFlow')}</Button>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {isFlowDeleteDialogOpen && activeFlow && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center">
          <div className="fixed inset-0 bg-black/50" onClick={() => setIsFlowDeleteDialogOpen(false)} />
          <div className="relative bg-white rounded-lg p-6 w-full max-w-md mx-4 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="flex-shrink-0 w-10 h-10 bg-red-100 rounded-full flex items-center justify-center"><AlertTriangle className="h-5 w-5 text-red-600" /></div>
              <div><h3 className="text-lg font-semibold text-gray-900">{t('links.deleteFlow')}</h3><p className="text-sm text-gray-500">{t('links.deleteFlowConfirm', { name: activeFlow.flow.name })}</p></div>
            </div>
            {activeFlow.links.length > 0 && (
              <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-md">
                <p className="text-sm text-amber-800 mb-3">{t('links.flowContainsLinks', { count: activeFlow.links.length })}</p>
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {activeFlow.links.map(link => (
                    <label key={link.id} className="flex items-center">
                      <input type="checkbox" checked={flowLinksToKeep.has(link.id)} onChange={(e) => { const newSet = new Set(flowLinksToKeep); if (e.target.checked) newSet.add(link.id); else newSet.delete(link.id); setFlowLinksToKeep(newSet); }} className="mr-2" />
                      <span className="text-sm text-gray-700">{link.name}</span>
                    </label>
                  ))}
                </div>
                <div className="mt-3 pt-3 border-t border-amber-200"><p className="text-xs text-amber-700">{t('links.flowKeepSelected')}<br />{t('links.flowKeepUnselected')}</p></div>
              </div>
            )}
            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsFlowDeleteDialogOpen(false)}>{t('links.cancel')}</Button>
              <Button variant="destructive" onClick={handleConfirmDeleteFlow}>{t('links.confirmDelete')}</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
