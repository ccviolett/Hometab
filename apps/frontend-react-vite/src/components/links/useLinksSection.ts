import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { GroupedLinks, Link, LinkFlow, LinkFlowWithLinks, LinkGroup } from '@/types/link'
import IconService, { type IconCheckResult } from '@/lib/iconService'
import type { LinkFormData } from './LinkEditDialog'
import { apiClient } from '@/lib/api-client'
import { useCreateLink, useDeleteLink, useLinksByGroup, useUpdateLink } from '@/queries/use-links'
import { useCreateLinkGroup, useDeleteLinkGroup, useUpdateLinkGroup } from '@/queries/use-link-groups'
import { useCreateLinkFlow, useDeleteLinkFlow, useRemoveLinkFromFlow, useUpdateFlowLinkOrder } from '@/queries/use-link-flows'

interface IconConflictState {
  host: string
  currentIconUrl: string
  pendingIconUrl: string
}

export function useLinksSection() {
  const { t } = useTranslation();
  const [groupedLinks, setGroupedLinks] = useState<GroupedLinks[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isGroupDialogOpen, setIsGroupDialogOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<LinkGroup | null>(null);
  const [isLinkDialogOpen, setIsLinkDialogOpen] = useState(false);
  const [editingLink, setEditingLink] = useState<Link | null>(null);
  const [selectedFlowForNewLink, setSelectedFlowForNewLink] = useState<LinkFlow | null>(null);
  const [isFlowDialogOpen, setIsFlowDialogOpen] = useState(false);
  const [flowDialogGroupId, setFlowDialogGroupId] = useState<string>('');
  const [flowNameInput, setFlowNameInput] = useState('');
  const [isFlowDetailOpen, setIsFlowDetailOpen] = useState(false);
  const [activeFlow, setActiveFlow] = useState<LinkFlowWithLinks | null>(null);
  const [flowLinksToKeep, setFlowLinksToKeep] = useState<Set<string>>(new Set());
  const [isFlowDeleteDialogOpen, setIsFlowDeleteDialogOpen] = useState(false);
  const [isDeleteGroupDialogOpen, setIsDeleteGroupDialogOpen] = useState(false);
  const [groupToDelete, setGroupToDelete] = useState<LinkGroup | null>(null);
  const [deleteOption, setDeleteOption] = useState<'move' | 'delete'>('move');
  const [selectedGroupForNewLink, setSelectedGroupForNewLink] = useState<LinkGroup | null>(null);
  const [sortingGroupId, setSortingGroupId] = useState<string | null>(null);
  const [isDeleteLinkDialogOpen, setIsDeleteLinkDialogOpen] = useState(false);
  const [linkToDelete, setLinkToDelete] = useState<Link | null>(null);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [sortingMode, setSortingMode] = useState(false);
  const [iconConflict, setIconConflict] = useState<IconConflictState | null>(null);
  const [iconVersion, setIconVersion] = useState(0);
  const { data: queriedGroupedLinks = [], isLoading, error: queryError, refetch } = useLinksByGroup();
  const createLink = useCreateLink();
  const updateLink = useUpdateLink();
  const deleteLink = useDeleteLink();
  const createLinkGroup = useCreateLinkGroup();
  const updateLinkGroup = useUpdateLinkGroup();
  const deleteLinkGroup = useDeleteLinkGroup();
  const createLinkFlow = useCreateLinkFlow();
  const deleteLinkFlow = useDeleteLinkFlow();
  const removeLinkFromFlow = useRemoveLinkFromFlow();
  const updateFlowLinkOrder = useUpdateFlowLinkOrder();
  const updateLinkOrder = (id: string, groupId: string, orderIndex: number) =>
    updateLink.mutateAsync({ id, data: { group_id: groupId, order_index: orderIndex } });

  useEffect(() => {
    try {
      const saved = localStorage.getItem('collapsedLinkGroups');
      if (saved) { setCollapsedGroups(new Set(JSON.parse(saved))); }
    } catch (error) { console.error('Failed to load collapsed groups:', error); }
  }, []);

  useEffect(() => {
    queueMicrotask(() => {
      setGroupedLinks(queriedGroupedLinks || []);
      setError(queryError ? t('links.loadError') : null);
    });
  }, [queriedGroupedLinks, queryError, t]);

  useEffect(() => {
    try { localStorage.setItem('collapsedLinkGroups', JSON.stringify(Array.from(collapsedGroups))); }
    catch (error) { console.error('Failed to save collapsed groups:', error); }
  }, [collapsedGroups]);

  const loadGroupedLinks = async (): Promise<GroupedLinks[]> => {
    try {
      const queryResult = await refetch();
      if (queryResult.error) throw queryResult.error;
      const data = queryResult.data || [];
      setGroupedLinks(data || []);
      setError(null);
      return data;
    } catch (error) {
      console.error('Failed to load grouped links:', error);
      setError(t('links.loadError'));
      setGroupedLinks([]);
      return [];
    }
  };

  const handleCreateGroup = async (groupData: Omit<LinkGroup, 'id' | 'created_at' | 'updated_at'>) => {
    try {
      if (editingGroup) {
        await updateLinkGroup.mutateAsync({ id: editingGroup.id, data: groupData });
      } else {
        const nonUngrouped = (groupedLinks || []).filter(g => g.group.name !== '未分组').map(g => g.group.order_index);
        const maxOrderIndex = nonUngrouped.length > 0 ? Math.max(...nonUngrouped) : 0;
        await createLinkGroup.mutateAsync({ ...groupData, order_index: Math.max(0, maxOrderIndex + 10) });
      }
      await loadGroupedLinks();
      setEditingGroup(null);
    } catch (error) { console.error('Failed to create/update group:', error); }
  };

  const handleEditGroup = (group: LinkGroup) => { setEditingGroup(group); setIsGroupDialogOpen(true); };
  const handleDeleteGroup = (group: LinkGroup) => { setGroupToDelete(group); setIsDeleteGroupDialogOpen(true); };

  const handleConfirmDeleteGroup = async () => {
    if (!groupToDelete) return;
    try {
      if (deleteOption === 'delete') {
        const groupData = groupedLinks.find(g => g.group.id === groupToDelete.id);
        if (groupData) { for (const link of groupData.links) { await deleteLink.mutateAsync(link.id); } }
      }
      await deleteLinkGroup.mutateAsync(groupToDelete.id);
      await loadGroupedLinks();
      setIsDeleteGroupDialogOpen(false);
      setGroupToDelete(null);
    } catch (error) { console.error('Failed to delete group:', error); }
  };

  const handleEditLink = (link: Link) => { setEditingLink(link); setIsLinkDialogOpen(true); };
  const handleDeleteLink = (link: Link) => { setLinkToDelete(link); setIsDeleteLinkDialogOpen(true); };

  const handleCopyLink = async (link: Link) => {
    try {
      const copied = await createLink.mutateAsync({ name: t('links.copyName', { name: link.name }), url: link.url, group_id: link.group_id || '', flow_id: link.flow_id, order_index: link.order_index + 1 });
      void checkLinkIcon(copied.url);
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow && isFlowDetailOpen && link.flow_id === activeFlow.flow.id) {
        const updatedFlow = newGroupedLinks.find(g => g.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === activeFlow.flow.id);
        if (updatedFlow) setActiveFlow(updatedFlow);
      }
    } catch (error) { console.error('Failed to copy link:', error); }
  };

  const handleConfirmDeleteLink = async () => {
    if (!linkToDelete) return;
    try {
      await deleteLink.mutateAsync(linkToDelete.id);
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow && isFlowDetailOpen && linkToDelete.flow_id === activeFlow.flow.id) {
        const updatedFlow = newGroupedLinks.find(g => g.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === activeFlow.flow.id);
        if (updatedFlow) setActiveFlow(updatedFlow);
      }
      setIsDeleteLinkDialogOpen(false);
      setLinkToDelete(null);
    } catch (error) { console.error('Failed to delete link:', error); }
  };

  const handleAddLink = () => { setEditingLink(null); setSelectedGroupForNewLink(null); setSelectedFlowForNewLink(null); setIsLinkDialogOpen(true); };
  const handleAddLinkToGroup = (group: LinkGroup) => { setEditingLink(null); setSelectedGroupForNewLink(group); setSelectedFlowForNewLink(null); setIsLinkDialogOpen(true); };
  const handleAddLinkToFlow = (group: LinkGroup, flow: LinkFlow) => { setEditingLink(null); setSelectedGroupForNewLink(group); setSelectedFlowForNewLink(flow); setIsLinkDialogOpen(true); };

  const handleFlowCreated = async () => {
    const groupId = flowDialogGroupId;
    const flowName = flowNameInput.trim();
    if (!groupId || !flowName) return;
    try {
      const groupFlows = groupedLinks.find(g => g.group.id === groupId)?.flows || [];
      const maxOrder = groupFlows.length > 0 ? Math.max(...groupFlows.map(f => f.flow.order_index)) : -10;
      await createLinkFlow.mutateAsync({ group_id: groupId, name: flowName, order_index: maxOrder + 10 });
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to create flow:', error); }
    finally { setIsFlowDialogOpen(false); }
  };

  const handleOpenFlowDetail = (flow: LinkFlowWithLinks) => { setActiveFlow(flow); setFlowLinksToKeep(new Set(flow.links.map(link => link.id))); setIsFlowDetailOpen(true); };

  const handleExecuteFlow = (flow: LinkFlowWithLinks) => {
    const sortedLinks = [...flow.links].sort((a, b) => a.order_index - b.order_index);
    sortedLinks.forEach((link) => { window.open(link.url, '_blank', 'noopener,noreferrer'); });
  };

  const moveFlowLink = async (flowId: string, links: Link[], linkId: string, direction: 'left' | 'right') => {
    const sortedLinks = [...links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = sortedLinks.findIndex(link => link.id === linkId);
    if (currentIndex === -1) return;
    const targetIndex = direction === 'left' ? currentIndex - 1 : currentIndex + 1;
    if (targetIndex < 0 || targetIndex >= sortedLinks.length) return;
    const currentLink = sortedLinks[currentIndex]; const targetLink = sortedLinks[targetIndex];
    try {
      await updateFlowLinkOrder.mutateAsync({ flowId, linkId: currentLink.id, orderIndex: targetLink.order_index });
      await updateFlowLinkOrder.mutateAsync({ flowId, linkId: targetLink.id, orderIndex: currentLink.order_index });
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow) {
        const updatedFlow = newGroupedLinks.find(group => group.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === flowId);
        if (updatedFlow) setActiveFlow(updatedFlow);
      }
    } catch (error) { console.error('Failed to move flow link:', error); }
  };

  const moveFlowLinkToStart = async (flowId: string, links: Link[], linkId: string) => {
    const sortedLinks = [...links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = sortedLinks.findIndex(link => link.id === linkId);
    if (currentIndex === -1 || currentIndex === 0) return;
    try {
      const newOrderIndexes = sortedLinks.map((link, index) => {
        if (link.id === linkId) return 0;
        else if (index < currentIndex) return (index + 1) * 10;
        else return index * 10;
      });
      await Promise.all(sortedLinks.map((link, index) => updateFlowLinkOrder.mutateAsync({ flowId, linkId: link.id, orderIndex: newOrderIndexes[index] })));
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow) { const updatedFlow = newGroupedLinks.find(group => group.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === flowId); if (updatedFlow) setActiveFlow(updatedFlow); }
    } catch (error) { console.error('Failed to move flow link to start:', error); }
  };

  const moveFlowLinkToEnd = async (flowId: string, links: Link[], linkId: string) => {
    const sortedLinks = [...links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = sortedLinks.findIndex(link => link.id === linkId);
    if (currentIndex === -1 || currentIndex === sortedLinks.length - 1) return;
    try {
      const newOrderIndexes = sortedLinks.map((link, index) => {
        if (link.id === linkId) return (sortedLinks.length - 1) * 10;
        else if (index > currentIndex) return (index - 1) * 10;
        else return index * 10;
      });
      await Promise.all(sortedLinks.map((link, index) => updateFlowLinkOrder.mutateAsync({ flowId, linkId: link.id, orderIndex: newOrderIndexes[index] })));
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow) { const updatedFlow = newGroupedLinks.find(group => group.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === flowId); if (updatedFlow) setActiveFlow(updatedFlow); }
    } catch (error) { console.error('Failed to move flow link to end:', error); }
  };

  const handleRemoveLinkFromFlow = async (flowId: string, linkId: string) => {
    try {
      await removeLinkFromFlow.mutateAsync({ flowId, linkId });
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow) {
        const updatedFlow = newGroupedLinks.find(group => group.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === flowId);
        if (updatedFlow) { setActiveFlow(updatedFlow); setFlowLinksToKeep(new Set(updatedFlow.links.map(link => link.id))); }
        else { setIsFlowDetailOpen(false); }
      }
    } catch (error) { console.error('Failed to remove link from flow:', error); }
  };

  const handleConfirmDeleteFlow = async () => {
    if (!activeFlow) return;
    try {
      await deleteLinkFlow.mutateAsync({ id: activeFlow.flow.id, linkIdsToKeep: Array.from(flowLinksToKeep) });
      await loadGroupedLinks();
      setIsFlowDeleteDialogOpen(false); setIsFlowDetailOpen(false); setActiveFlow(null);
    } catch (error) { console.error('Failed to delete flow:', error); }
  };

  const checkLinkIcon = async (url: string) => {
    try {
      const result: IconCheckResult = await IconService.checkIcon(url);
      if (result.status === 'conflict' && result.current_icon_url && result.pending_icon_url) {
        setIconConflict({
          host: result.host,
          currentIconUrl: result.current_icon_url,
          pendingIconUrl: result.pending_icon_url,
        });
      } else if (result.status === 'ready' || result.status === 'unchanged') {
        setIconVersion(v => v + 1);
      }
    } catch (error) {
      console.error('Failed to check link icon:', error);
    }
  };

  const handleChooseIcon = async (choice: 'current' | 'new') => {
    if (!iconConflict) return;
    try {
      await IconService.chooseIcon(iconConflict.host, choice);
      setIconVersion(v => v + 1);
      setIconConflict(null);
    } catch (error) {
      console.error('Failed to choose icon:', error);
    }
  };

  const reorderGroups = async (ids: string[]) => {
    try { await apiClient.put('api/link-groups/order', { ids }); await loadGroupedLinks(); }
    catch (error) { console.error('Failed to reorder groups:', error); setError(t('links.sortFailed')); await loadGroupedLinks(); }
  };

  const reorderGroupLinks = async (groupId: string, ids: string[]) => {
    try { await apiClient.put(`api/link-groups/${groupId}/links/order`, { ids }); await loadGroupedLinks(); }
    catch (error) { console.error('Failed to reorder links:', error); setError(t('links.sortFailed')); await loadGroupedLinks(); }
  };

  const reorderFlowLinks = async (flowId: string, ids: string[]) => {
    try {
      await apiClient.put(`api/link-flows/${flowId}/links/order`, { ids });
      const next = await loadGroupedLinks();
      if (activeFlow) {
        const updated = next.flatMap((group) => group.flows).find((flow) => flow.flow.id === flowId);
        if (updated) setActiveFlow(updated);
      }
    } catch (error) { console.error('Failed to reorder flow links:', error); setError(t('links.sortFailed')); await loadGroupedLinks(); }
  };

  const handleToggleSorting = (groupId: string) => { setSortingGroupId(sortingGroupId === groupId ? null : groupId); };

  const handleMoveLink = async (linkId: string, direction: 'left' | 'right') => {
    if (!sortingGroupId) return;
    const groupData = groupedLinks.find(g => g.group.id === sortingGroupId);
    if (!groupData) return;
    const links = [...groupData.links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = links.findIndex(l => l.id === linkId);
    if (currentIndex === -1) return;
    const swapIndex = direction === 'left' ? currentIndex - 1 : currentIndex + 1;
    if (swapIndex < 0 || swapIndex >= links.length) return;
    const currentLink = links[currentIndex]; const swapLink = links[swapIndex];

    if (currentLink.order_index === swapLink.order_index) {
      const updates = links.map((link, index) => ({ link, newOrderIndex: index }));
      const temp = updates[currentIndex]; updates[currentIndex] = updates[swapIndex]; updates[swapIndex] = temp;
      try { for (const { link, newOrderIndex } of updates) { if (link.order_index !== newOrderIndex) await updateLinkOrder(link.id, sortingGroupId, newOrderIndex); } await loadGroupedLinks(); }
      catch (error) { console.error('Failed to update link order:', error); }
    } else {
      try { await updateLinkOrder(currentLink.id, sortingGroupId, swapLink.order_index); await updateLinkOrder(swapLink.id, sortingGroupId, currentLink.order_index); await loadGroupedLinks(); }
      catch (error) { console.error('Failed to update link order:', error); }
    }
  };

  const handleMoveLinkToStart = async (linkId: string) => {
    if (!sortingGroupId) return;
    const groupData = groupedLinks.find(g => g.group.id === sortingGroupId);
    if (!groupData) return;
    const links = [...groupData.links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = links.findIndex(l => l.id === linkId);
    if (currentIndex === -1 || currentIndex === 0) return;
    try {
      const updates = links.map((link, index) => { if (link.id === linkId) return { link, newOrderIndex: 0 }; else if (index < currentIndex) return { link, newOrderIndex: index + 1 }; else return { link, newOrderIndex: index }; });
      for (const { link, newOrderIndex } of updates) { await updateLinkOrder(link.id, sortingGroupId, newOrderIndex); }
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to move link to start:', error); }
  };

  const handleMoveLinkToEnd = async (linkId: string) => {
    if (!sortingGroupId) return;
    const groupData = groupedLinks.find(g => g.group.id === sortingGroupId);
    if (!groupData) return;
    const links = [...groupData.links].sort((a, b) => a.order_index - b.order_index);
    const currentIndex = links.findIndex(l => l.id === linkId);
    if (currentIndex === -1 || currentIndex === links.length - 1) return;
    try {
      const updates = links.map((link, index) => { if (link.id === linkId) return { link, newOrderIndex: links.length - 1 }; else if (index > currentIndex) return { link, newOrderIndex: index - 1 }; else return { link, newOrderIndex: index }; });
      for (const { link, newOrderIndex } of updates) { await updateLinkOrder(link.id, sortingGroupId, newOrderIndex); }
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to move link to end:', error); }
  };

  const handleSaveLink = async (linkData: LinkFormData) => {
    try {
      if (!linkData.group_id) {
        const ungroupedGroup = groupedLinks.find(g => g.group.name === '未分组');
        if (ungroupedGroup) linkData.group_id = ungroupedGroup.group.id;
      }
      if (editingLink) {
        const isChangingGroup = editingLink.group_id !== linkData.group_id;
        const isChangingFlow = editingLink.flow_id !== linkData.flow_id;
        const shouldCheckIcon = editingLink.url !== linkData.url;
        if (isChangingGroup || isChangingFlow) {
          let maxOrder = -1;
          if (linkData.flow_id) { const targetFlow = groupedLinks.find(g => g.group.id === linkData.group_id)?.flows.find(f => f.flow.id === linkData.flow_id); if (targetFlow && targetFlow.links.length > 0) maxOrder = Math.max(...targetFlow.links.map(link => link.order_index)); }
          else { const targetGroup = groupedLinks.find(g => g.group.id === linkData.group_id); if (targetGroup && targetGroup.links.length > 0) maxOrder = Math.max(...targetGroup.links.map(link => link.order_index)); }
          await updateLink.mutateAsync({ id: editingLink.id, data: { ...linkData, order_index: maxOrder + 1 } });
        } else { await updateLink.mutateAsync({ id: editingLink.id, data: linkData }); }
        if (shouldCheckIcon) void checkLinkIcon(linkData.url);
      } else {
        let maxOrder = -1;
        if (linkData.flow_id) { const targetFlow = groupedLinks.find(g => g.group.id === linkData.group_id)?.flows.find(f => f.flow.id === linkData.flow_id); if (targetFlow && targetFlow.links.length > 0) maxOrder = Math.max(...targetFlow.links.map(link => link.order_index)); }
        else { const targetGroup = groupedLinks.find(g => g.group.id === linkData.group_id); if (targetGroup && targetGroup.links.length > 0) maxOrder = Math.max(...targetGroup.links.map(link => link.order_index)); }
        const created = await createLink.mutateAsync({ ...linkData, order_index: maxOrder + 1 });
        void checkLinkIcon(created.url);
      }
      const newGroupedLinks = await loadGroupedLinks();
      if (activeFlow && isFlowDetailOpen) {
        if (linkData.flow_id) { const updatedFlow = newGroupedLinks.find(g => g.group.id === linkData.group_id)?.flows.find(f => f.flow.id === linkData.flow_id); if (updatedFlow) setActiveFlow(updatedFlow); }
        else if (editingLink?.flow_id === activeFlow.flow.id) { const updatedFlow = newGroupedLinks.find(g => g.group.id === activeFlow.flow.group_id)?.flows.find(f => f.flow.id === activeFlow.flow.id); if (updatedFlow) setActiveFlow(updatedFlow); }
      }
      setIsLinkDialogOpen(false); setEditingLink(null);
    } catch (error) { console.error('Failed to save link:', error); }
  };

  const toggleGroupCollapse = (groupId: string) => {
    setCollapsedGroups(prev => { const newCollapsed = new Set(prev); if (newCollapsed.has(groupId)) newCollapsed.delete(groupId); else newCollapsed.add(groupId); return newCollapsed; });
  };

  const swapGroupOrder = async (groupId: string, direction: 'up' | 'down') => {
    const currentIndex = groupedLinks.findIndex(g => g.group.id === groupId);
    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;
    if (targetIndex < 0 || targetIndex >= groupedLinks.length) return;
    try {
      const newGroupedLinks = [...groupedLinks]; [newGroupedLinks[currentIndex], newGroupedLinks[targetIndex]] = [newGroupedLinks[targetIndex], newGroupedLinks[currentIndex]];
      setGroupedLinks(newGroupedLinks);
      await Promise.all(newGroupedLinks.map((group, index) => updateLinkGroup.mutateAsync({ id: group.group.id, data: { order_index: index * 10 } })));
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to update group order:', error); setError(t('links.updateGroupOrderError')); }
  };

  const moveGroupToStart = async (groupId: string) => {
    const currentIndex = groupedLinks.findIndex(g => g.group.id === groupId);
    if (currentIndex === 0) return;
    try {
      const newGroupedLinks = [...groupedLinks]; const [movedGroup] = newGroupedLinks.splice(currentIndex, 1); newGroupedLinks.unshift(movedGroup);
      setGroupedLinks(newGroupedLinks);
      await Promise.all(newGroupedLinks.map((group, index) => updateLinkGroup.mutateAsync({ id: group.group.id, data: { order_index: index * 10 } })));
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to move group to start:', error); setError(t('links.moveGroupError')); }
  };

  const moveGroupToEnd = async (groupId: string) => {
    const currentIndex = groupedLinks.findIndex(g => g.group.id === groupId);
    if (currentIndex === groupedLinks.length - 1) return;
    try {
      const newGroupedLinks = [...groupedLinks]; const [movedGroup] = newGroupedLinks.splice(currentIndex, 1); newGroupedLinks.push(movedGroup);
      setGroupedLinks(newGroupedLinks);
      await Promise.all(newGroupedLinks.map((group, index) => updateLinkGroup.mutateAsync({ id: group.group.id, data: { order_index: index * 10 } })));
      await loadGroupedLinks();
    } catch (error) { console.error('Failed to move group to end:', error); setError(t('links.moveGroupError')); }
  };


  return { t, groupedLinks, error, isGroupDialogOpen, setIsGroupDialogOpen, editingGroup, isLinkDialogOpen, setIsLinkDialogOpen, editingLink, selectedFlowForNewLink, isFlowDialogOpen, setIsFlowDialogOpen, flowDialogGroupId, setFlowDialogGroupId, flowNameInput, setFlowNameInput, isFlowDetailOpen, setIsFlowDetailOpen, activeFlow, flowLinksToKeep, setFlowLinksToKeep, isFlowDeleteDialogOpen, setIsFlowDeleteDialogOpen, isDeleteGroupDialogOpen, setIsDeleteGroupDialogOpen, groupToDelete, deleteOption, setDeleteOption, selectedGroupForNewLink, sortingGroupId, isDeleteLinkDialogOpen, setIsDeleteLinkDialogOpen, linkToDelete, setLinkToDelete, collapsedGroups, sortingMode, setSortingMode, iconConflict, setIconConflict, iconVersion, isLoading, handleCreateGroup, handleEditGroup, handleDeleteGroup, handleConfirmDeleteGroup, handleEditLink, handleDeleteLink, handleCopyLink, handleConfirmDeleteLink, handleAddLink, handleAddLinkToGroup, handleAddLinkToFlow, handleFlowCreated, handleOpenFlowDetail, handleExecuteFlow, moveFlowLink, moveFlowLinkToStart, moveFlowLinkToEnd, handleRemoveLinkFromFlow, handleConfirmDeleteFlow, handleChooseIcon, handleToggleSorting, handleMoveLink, handleMoveLinkToStart, handleMoveLinkToEnd, handleSaveLink, toggleGroupCollapse, swapGroupOrder, moveGroupToStart, moveGroupToEnd, reorderGroups, reorderGroupLinks, reorderFlowLinks }
}
