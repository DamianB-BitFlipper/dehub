'use client';

import type { DiffIndicators } from '@pierre/diffs';
import {
  IconBranch,
  IconCodeStyleBars,
  IconCollapsedRow,
  IconDiffSplit,
  IconDiffUnified,
  IconExpandAll,
  IconEyeSlash,
  IconGearFill,
  IconSearch,
  IconSidebarLeft,
  IconSidebarLeftOpen,
  IconSymbolDiffstat,
  IconXSquircle,
} from '@pierre/icons';
import { FileTree } from '@pierre/trees';
import { useFileTreeSearch } from '@pierre/trees/react';
import {
  type Dispatch,
  memo,
  type ReactNode,
  type RefObject,
  type SetStateAction,
  useCallback,
  useEffect,
  useState,
} from 'react';

import { CodeViewDiffStats } from './CodeViewDiffStats';
import { CodeViewFileTree } from './CodeViewFileTree';
import type {
  CodeViewDiffStats as CodeViewDiffStatsData,
  CodeViewFileTreeSource,
  CodeViewSavedCommentItem,
} from './types';
import { WorkerPoolStatus } from './WorkerPoolStatus';
import { useOverflowMarquee } from '../useOverflowMarquee';
import { Button } from '@/components/ui/button';
import { ButtonGroup, ButtonGroupItem } from '@/components/ui/button-group';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';
import { cn } from '@/lib/utils';

type SidebarStatusPanel = 'diffStats' | 'systemMonitor';
type SidebarTab = 'files';

const MOBILE_MEDIA_QUERY = '(max-width: 767px)';
const SETTING_ROW_CLASS =
  'w-full flex cursor-pointer items-center justify-between gap-4 px-2 py-1.5 text-sm';

interface CodeViewSidebarProps {
  className?: string;
  prTitle?: string;
  baseRefName?: string;
  headRefName?: string;
  commentSections: readonly CodeViewSavedCommentItem[];
  collapseMode: 'expanded' | 'collapsed';
  diffIndicators: DiffIndicators;
  diffStyle: 'split' | 'unified';
  diffStats: CodeViewDiffStatsData | null;
  lineNumbers: boolean;
  mobileOverlayOpen?: boolean;
  onMobileClose(): void;
  onSelectItem(itemId: string): void;
  onToggleCollapseMode(): void;
  onToggleSidebarCollapsed(): void;
  overflow: 'wrap' | 'scroll';
  scrollRef: RefObject<HTMLDivElement | null>;
  setDiffIndicators: Dispatch<SetStateAction<DiffIndicators>>;
  setDiffStyle: Dispatch<SetStateAction<'split' | 'unified'>>;
  setLineNumbers: Dispatch<SetStateAction<boolean>>;
  setOverflow: Dispatch<SetStateAction<'wrap' | 'scroll'>>;
  setShowBackgrounds: Dispatch<SetStateAction<boolean>>;
  showBackgrounds: boolean;
  sidebarCollapsed: boolean;
  source: CodeViewFileTreeSource;
  streaming: boolean;
}

export const CodeViewSidebar = memo(function CodeViewSidebar({
  className,
  prTitle,
  baseRefName,
  headRefName,
  commentSections,
  collapseMode,
  diffIndicators,
  diffStyle,
  diffStats,
  lineNumbers,
  mobileOverlayOpen = false,
  onMobileClose,
  onSelectItem,
  onToggleCollapseMode,
  onToggleSidebarCollapsed,
  overflow,
  scrollRef,
  setDiffIndicators,
  setDiffStyle,
  setLineNumbers,
  setOverflow,
  setShowBackgrounds,
  showBackgrounds,
  sidebarCollapsed,
  source,
  streaming,
}: CodeViewSidebarProps) {
  void commentSections;
  const activeTab: SidebarTab = 'files';
  const [activeStatusPanel, setActiveStatusPanel] =
    useState<SidebarStatusPanel | null>('diffStats');
  const [fileTreeModel, setFileTreeModel] = useState<FileTree | null>(null);
  const handleModelReady = useCallback((model: FileTree | null) => {
    setFileTreeModel(model);
  }, []);
  const toggleStatusPanel = useCallback((panel: SidebarStatusPanel) => {
    setActiveStatusPanel((current) => (current === panel ? null : panel));
  }, []);

  useEffect(() => {
    if (mobileOverlayOpen && window.matchMedia(MOBILE_MEDIA_QUERY).matches) {
      setActiveStatusPanel(null);
    }
  }, [mobileOverlayOpen]);

  useEffect(() => {
    if (!mobileOverlayOpen || !window.matchMedia(MOBILE_MEDIA_QUERY).matches) {
      return undefined;
    }

    const { body, documentElement } = document;
    const codeViewScroll = scrollRef.current;
    const previousBodyOverflow = body.style.overflow;
    const previousRootOverscrollBehavior =
      documentElement.style.overscrollBehavior;
    const previousCodeViewOverflow = codeViewScroll?.style.overflow;

    body.style.overflow = 'hidden';
    documentElement.style.overscrollBehavior = 'none';
    if (codeViewScroll != null) {
      codeViewScroll.style.overflow = 'hidden';
    }

    return () => {
      body.style.overflow = previousBodyOverflow;
      documentElement.style.overscrollBehavior = previousRootOverscrollBehavior;
      if (codeViewScroll != null) {
        codeViewScroll.style.overflow = previousCodeViewOverflow ?? '';
      }
    };
  }, [mobileOverlayOpen, scrollRef]);

  return (
    <>
      <button
        type="button"
        aria-hidden={!mobileOverlayOpen}
        aria-label="Close file tree"
        tabIndex={mobileOverlayOpen ? 0 : -1}
        className={cn(
          'z-20 cursor-default bg-background/60 backdrop-blur-xs transition-opacity [grid-column:1/-1] [grid-row:1/-1] md:hidden',
          mobileOverlayOpen
            ? 'pointer-events-auto opacity-100'
            : 'pointer-events-none opacity-0'
        )}
        onClick={onMobileClose}
      />
      <SidebarWrapper
        className={className}
        mobileOverlayOpen={mobileOverlayOpen}
        sidebarCollapsed={sidebarCollapsed}
      >
        <div
          className={cn(
            'flex items-center gap-3 px-4 pt-5 pb-2 md:px-3 md:pt-0.5 md:pb-0',
            sidebarCollapsed && 'justify-start px-2 pb-0 md:px-2'
          )}
        >
          <div
            className={cn(
              'mr-auto flex min-w-0 items-center gap-3 md:gap-2',
              sidebarCollapsed && 'mr-0'
            )}
          >
            <Button
              type="button"
              variant="ghost"
              size="icon-md"
              aria-label={sidebarCollapsed ? 'Expand file tree' : 'Collapse file tree'}
              aria-pressed={!sidebarCollapsed}
              title={sidebarCollapsed ? 'Expand file tree' : 'Collapse file tree'}
              className={cn(
                'border border-[var(--color-border-opaque)] bg-background/20 shadow-none hover:bg-accent/60 hover:text-accent-foreground',
                sidebarCollapsed && 'bg-muted/70'
              )}
              onClick={onToggleSidebarCollapsed}
            >
              {sidebarCollapsed ? (
                <IconSidebarLeftOpen className="size-4 md:size-3" />
              ) : (
                <IconSidebarLeft className="size-4 md:size-3" />
              )}
            </Button>
            {!sidebarCollapsed && (
              <SidebarDiffControls
                collapseMode={collapseMode}
                diffIndicators={diffIndicators}
                diffStyle={diffStyle}
                lineNumbers={lineNumbers}
                overflow={overflow}
                onToggleCollapseMode={onToggleCollapseMode}
                setDiffIndicators={setDiffIndicators}
                setDiffStyle={setDiffStyle}
                setLineNumbers={setLineNumbers}
                setOverflow={setOverflow}
                setShowBackgrounds={setShowBackgrounds}
                showBackgrounds={showBackgrounds}
              />
            )}
          </div>
          {!sidebarCollapsed && activeTab === 'files' && fileTreeModel != null && (
            <FileTreeSearchToggle model={fileTreeModel} />
          )}
          {onMobileClose != null && (
            <Button
              variant="ghost"
              size="icon-only"
              className="md:hidden"
              aria-label="Close file tree"
              onClick={onMobileClose}
            >
              <IconXSquircle className="size-4 md:size-3" />
            </Button>
          )}
        </div>
        {!sidebarCollapsed && (
          <SidebarDiffMeta
            prTitle={prTitle}
            baseRefName={baseRefName}
            headRefName={headRefName}
          />
        )}
        <div className={cn('mt-3 min-h-0 flex-1', sidebarCollapsed && 'hidden')}>
          <div
            role="region"
            aria-label="Files"
            hidden={activeTab !== 'files'}
            className="h-full min-h-0"
          >
            <CodeViewFileTree
              source={source}
              onModelReady={handleModelReady}
              onSelectItem={onSelectItem}
            />
          </div>
        </div>
        {!sidebarCollapsed && (
          <>
            <CodeViewDiffStats
              expanded={activeStatusPanel === 'diffStats'}
              onToggle={() => toggleStatusPanel('diffStats')}
              stats={diffStats}
              streaming={streaming}
            />
            <WorkerPoolStatus
              expanded={activeStatusPanel === 'systemMonitor'}
              onToggle={() => toggleStatusPanel('systemMonitor')}
              scrollRef={scrollRef}
            />
          </>
        )}
      </SidebarWrapper>
    </>
  );
});

interface SidebarDiffMetaProps {
  prTitle?: string;
  baseRefName?: string;
  headRefName?: string;
}

function SidebarDiffMeta({
  prTitle,
  baseRefName,
  headRefName,
}: SidebarDiffMetaProps) {
  const title = (prTitle ?? '').trim();
  const base = (baseRefName ?? '').trim();
  const head = (headRefName ?? '').trim();
  const hasBranches = base !== '' || head !== '';
  if (title === '' && !hasBranches) {
    return null;
  }

  return (
    <div className="mt-3 flex flex-col gap-2 border-b border-[var(--color-border-opaque)] px-4 pb-3 md:px-3">
      {title !== '' && (
        <MarqueeText
          value={title}
          className="text-sm font-semibold text-foreground"
        />
      )}
      {hasBranches && (
        <div className="flex min-w-0 flex-col gap-1">
          <BranchField value={head || '?'} />
          <div className="flex justify-center" aria-hidden>
            <span className="rounded-full border border-[var(--color-border-opaque)] bg-background/40 px-2 py-0.5 text-[10px] font-medium tracking-wide text-muted-foreground/70 uppercase">
              into
            </span>
          </div>
          <BranchField value={base || '?'} />
        </div>
      )}
    </div>
  );
}

function BranchField({ value }: { value: string }) {
  return (
    <div className="flex min-w-0 items-center gap-1.5 rounded-md border border-[var(--color-border-opaque)] bg-muted/35 px-2 py-1">
      <IconBranch className="size-3 shrink-0 text-muted-foreground/70" />
      <MarqueeText
        value={value}
        className="font-mono text-xs text-foreground"
      />
    </div>
  );
}

function MarqueeText({
  value,
  className,
}: {
  value: string;
  className?: string;
}) {
  const { containerRef, trackRef, overflowing, style } =
    useOverflowMarquee(value);

  return (
    <div
      ref={containerRef}
      className={cn('marquee-clip min-w-0 overflow-hidden', className)}
      style={style}
      title={value}
      tabIndex={overflowing ? 0 : undefined}
    >
      <span
        ref={trackRef}
        className={cn(
          'marquee-track',
          overflowing && 'marquee-track--overflow'
        )}
      >
        {value}
      </span>
    </div>
  );
}

interface SidebarWrapperProps {
  children: ReactNode;
  className?: string;
  mobileOverlayOpen: boolean;
  sidebarCollapsed: boolean;
}

function SidebarWrapper({
  children,
  className,
  mobileOverlayOpen,
  sidebarCollapsed,
}: SidebarWrapperProps) {
  return (
    <div
      className={cn(
        className,
        'bg-[var(--diffshub-sidebar-bg)] contain-strict z-30 flex h-full min-h-0 flex-col transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] will-change-transform motion-reduce:transition-none md:z-auto md:translate-y-0 md:will-change-auto',
        sidebarCollapsed && 'md:overflow-hidden',
        mobileOverlayOpen
          ? 'pointer-events-auto translate-y-0 overflow-hidden rounded-t-xl shadow-[0_0_0_1px_var(--color-border-opaque),_0_16px_32px_rgb(0_0_0_/0.25)] md:h-full md:overflow-visible md:rounded-none md:border-0 md:shadow-none'
          : 'pointer-events-none translate-y-[calc(100%+1.5rem)] overflow-hidden rounded-xl md:pointer-events-auto md:h-full md:overflow-visible md:rounded-none pt-3 border-r border-[var(--color-border-opaque)]'
      )}
    >
      {children}
    </div>
  );
}

interface SidebarDiffControlsProps {
  collapseMode: 'expanded' | 'collapsed';
  diffIndicators: DiffIndicators;
  diffStyle: 'split' | 'unified';
  lineNumbers: boolean;
  overflow: 'wrap' | 'scroll';
  onToggleCollapseMode(): void;
  setDiffIndicators: Dispatch<SetStateAction<DiffIndicators>>;
  setDiffStyle: Dispatch<SetStateAction<'split' | 'unified'>>;
  setLineNumbers: Dispatch<SetStateAction<boolean>>;
  setOverflow: Dispatch<SetStateAction<'wrap' | 'scroll'>>;
  setShowBackgrounds: Dispatch<SetStateAction<boolean>>;
  showBackgrounds: boolean;
}

function SidebarDiffControls({
  collapseMode,
  diffIndicators,
  diffStyle,
  lineNumbers,
  overflow,
  onToggleCollapseMode,
  setDiffIndicators,
  setDiffStyle,
  setLineNumbers,
  setOverflow,
  setShowBackgrounds,
  showBackgrounds,
}: SidebarDiffControlsProps) {
  return (
    <div className="flex items-center gap-2">
      <Button
        type="button"
        variant="ghost"
        size="icon-md"
        title={
          diffStyle === 'split' ? 'Switch to unified view' : 'Switch to split view'
        }
        className="hover:text-muted-foreground hidden hover:bg-transparent md:flex"
        onClick={() => setDiffStyle(diffStyle === 'split' ? 'unified' : 'split')}
      >
        {diffStyle === 'split' ? (
          <IconDiffSplit className="size-4 md:size-3" />
        ) : (
          <IconDiffUnified className="size-4 md:size-3" />
        )}
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="icon-md"
        aria-pressed={collapseMode === 'collapsed'}
        title={collapseMode === 'expanded' ? 'Collapse all files' : 'Expand all files'}
        className="hover:text-muted-foreground hover:bg-transparent"
        onClick={onToggleCollapseMode}
      >
        {collapseMode === 'expanded' ? (
          <IconExpandAll className="size-4 md:size-3" />
        ) : (
          <IconCollapsedRow className="size-4 md:size-3" />
        )}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon-md"
            className="hover:text-muted-foreground hover:bg-transparent"
          >
            <IconGearFill className="size-4 md:size-3" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-52">
          <DropdownMenuItem
            className="cursor-default p-0"
            onSelect={(event) => event.preventDefault()}
          >
            <label className={SETTING_ROW_CLASS}>
              <span className="min-w-0 flex-1">Backgrounds</span>
              <Switch checked={showBackgrounds} onCheckedChange={setShowBackgrounds} />
            </label>
          </DropdownMenuItem>
          <DropdownMenuItem
            className="cursor-default p-0"
            onSelect={(event) => event.preventDefault()}
          >
            <label className={SETTING_ROW_CLASS}>
              <span className="min-w-0 flex-1">Line numbers</span>
              <Switch checked={lineNumbers} onCheckedChange={setLineNumbers} />
            </label>
          </DropdownMenuItem>
          <DropdownMenuItem
            className="cursor-default p-0"
            onSelect={(event) => event.preventDefault()}
          >
            <label className={SETTING_ROW_CLASS}>
              <span className="min-w-0 flex-1">Word wrap</span>
              <Switch
                checked={overflow === 'wrap'}
                onCheckedChange={(checked) =>
                  setOverflow(checked ? 'wrap' : 'scroll')
                }
              />
            </label>
          </DropdownMenuItem>
          <DropdownMenuItem
            className="w-full px-2 focus:bg-transparent"
            onSelect={(event) => event.preventDefault()}
          >
            <span>Indicator style</span>
            <ButtonGroup
              className="ml-auto"
              value={diffIndicators}
              onValueChange={(value) => setDiffIndicators(value as DiffIndicators)}
            >
              <ButtonGroupItem value="bars" className="size-7 p-0">
                <IconCodeStyleBars className="size-3" />
              </ButtonGroupItem>
              <ButtonGroupItem value="classic" className="size-7 p-0">
                <IconSymbolDiffstat className="size-3" />
              </ButtonGroupItem>
              <ButtonGroupItem value="none" className="size-7 p-0">
                <IconEyeSlash className="size-3" />
              </ButtonGroupItem>
            </ButtonGroup>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

// Lives in its own component so we can call useFileTreeSearch only once we
// actually have a model; conditional hook calls aren't allowed in the parent.
function FileTreeSearchToggle({ model }: { model: FileTree }) {
  const search = useFileTreeSearch(model);
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-only"
      aria-label={search.isOpen ? 'Hide file search' : 'Show file search'}
      aria-pressed={search.isOpen}
      // Avoid focus moving to this button before click: the tree search input
      // closes on blur, so without preventDefault the blur runs first, then
      // click sees isOpen false and calls open() again.
      onPointerDown={(event) => event.preventDefault()}
      onClick={() => {
        if (search.isOpen) {
          search.close();
        } else {
          search.open();
        }
      }}
    >
      <IconSearch className="size-4 md:size-3" />
    </Button>
  );
}
