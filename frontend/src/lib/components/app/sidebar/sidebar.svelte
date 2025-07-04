<script lang="ts" module>
	import { onDestroy, onMount } from 'svelte';
	import type { model } from '$lib/wailsjs/go/models';
</script>

<script lang="ts">
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import Plus from 'lucide-svelte/icons/plus';
	import Settings from 'lucide-svelte/icons/settings';
	import Activity from 'lucide-svelte/icons/activity';
	import Table2 from 'lucide-svelte/icons/table-2';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import SettingsDialog from '$lib/components/app/sidebar/settings-dialog.svelte';
	import { type ComponentProps } from 'svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';

	let {
		ref = $bindable(null),
		tabID = $bindable(0),
		tabName = $bindable(''),
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();

	import {
		postgresConnectionsMap,
		connectionDatabasesMap,
		databasesMap,
		loadingMap,
		dbLoadingMap,
		activeDBs,
		suggestions,
		selectedDBDisplay,
		currentColor,
		activePoolID
	} from '$lib/state.svelte';

	import {
		EstablishPostgresConnection,
		EstablishPostgresDatabaseConnection,
		GetPostgresConnections,
		RefreshPostgresDatabase,
		TerminatePostgresDatabaseConnection
	} from '$lib/wailsjs/go/app/Connections';
	import Button from '$lib/components/ui/button/button.svelte';
	import { toast } from 'svelte-sonner';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { SaveActiveDBProps } from '$lib/wailsjs/go/app/Tabs';

	const refresh = () => {
		GetPostgresConnections()
			.then((connections) => {
				for (let connection of connections) {
					postgresConnectionsMap.set(connection.ID, connection);
				}
			})
			.catch((error) => {
				// Handle errors from the EstablishPostgresDatabaseConnection call
				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	};

	function establishDatabaseConnection(id: number, dbID: string) {
		let db = databasesMap.get(dbID);

		if (!db) {
			toast.error('Database not found', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		// Check if connection is already established
		if (db.IsActive) {
			dbLoadingMap.set(dbID, false);
			return;
		}

		dbLoadingMap.set(dbID, true);

		// Establish connection
		EstablishPostgresDatabaseConnection(id, dbID, db.Name)
			.then((db) => {
				dbLoadingMap.set(dbID, false);
				databasesMap.set(db.ID, db);

				$activeDBs.push(db);
				$suggestions.push(...db.Tables);
				$suggestions.push(...db.Columns);

				$selectedDBDisplay = db.PostgresConnectionName + ' - ' + db.Name;
				$currentColor = db.Colour;
				$activePoolID = db.PoolID;
				SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

				toast.success('Connected to ' + db.Name, {});
			})
			.catch((error) => {
				dbLoadingMap.set(dbID, false);

				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function establishConnection(id: number) {
		loadingMap.set(id, true);
		// Check if connection is already established
		if ((connectionDatabasesMap.get(id)?.length || 0) > 0) {
			loadingMap.set(id, false);
			return;
		}

		// Establish connection
		EstablishPostgresConnection(id)
			.then((dbs) => {
				for (let db of dbs) {
					databasesMap.set(db.ID, db);
					if (db.IsActive) {
						$activeDBs.push(db);
						$suggestions.push(...db.Tables);
						$suggestions.push(...db.Columns);

						// Set active DB
						$selectedDBDisplay = db.PostgresConnectionName + ' - ' + db.Name;
						$currentColor = db.Colour;
						$activePoolID = db.PoolID;
						SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

						toast.success('Connected to ' + db.Name, {});
					}
				}
				connectionDatabasesMap.set(
					id,
					dbs.map((db) => db.ID)
				);
				loadingMap.set(id, false);
			})
			.catch((error) => {
				loadingMap.set(id, false);
				// Handle errors from the EstablishPostgresDatabaseConnection call
				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function getColorClass(color: string): string {
		const colorMap: Record<string, string> = {
			'bg-purple-500': 'bg-purple-500',
			'bg-indigo-500': 'bg-indigo-500',
			'bg-emerald-500': 'bg-emerald-500',
			'bg-red-500': 'bg-red-500',
			'bg-blue-500': 'bg-blue-500',
			'bg-green-500': 'bg-green-500',
			'bg-yellow-500': 'bg-yellow-500',
			'bg-orange-500': 'bg-orange-500',
			'bg-pink-500': 'bg-pink-500'
		};
		return colorMap[color] || '';
	}

	function terminateDBConnection(dbID: string) {
		let db = databasesMap.get(dbID);

		if (!db) {
			toast.error('Database not found', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		if (!db.IsActive) {
			toast.error('Connection is not active', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		TerminatePostgresDatabaseConnection(db.PoolID)
			.then((success) => {
				if (success) {
					$activeDBs.splice(
						$activeDBs.findIndex((db) => db.ID === dbID),
						1
					); // Use splice to remove the item
					$suggestions.splice(
						$suggestions.findIndex((table) => db.Tables.includes(table)),
						db.Tables.length
					);
					$suggestions.splice(
						$suggestions.findIndex((column) => db.Columns.includes(column)),
						db.Columns.length
					);
					db.IsActive = false;
					db.Tables = [];
					db.Columns = [];
					databasesMap.delete(dbID);
					databasesMap.set(dbID, db);

					if ($activeDBs.length == 0) {
						$selectedDBDisplay = 'Connect to a database';
						$currentColor = '';
						$activePoolID = '';
					} else {
						// Set active DB
						$selectedDBDisplay = $activeDBs[0].PostgresConnectionName + ' - ' + $activeDBs[0].Name;
						$currentColor = $activeDBs[0].Colour;
						$activePoolID = $activeDBs[0].PoolID;
					}

					SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

					toast.success('Disconnected ' + db.Name, {});
				}
			})
			.catch((error) => {
				toast.error('Failed to disconnect', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function refreshDB(dbID: string) {
		dbLoadingMap.set(dbID, true);

		let db = databasesMap.get(dbID);
		if (!db) {
			dbLoadingMap.set(dbID, false);
			return;
		}

		// Establish connection
		RefreshPostgresDatabase(db.PostgresConnectionID, dbID, db.Name, db.PoolID)
			.then((db) => {
				dbLoadingMap.set(dbID, false);
				databasesMap.set(db.ID, db);
			})
			.catch((error) => {
				dbLoadingMap.set(dbID, false);

				toast.error('Failed to refresh', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}
</script>

<Sidebar.Root bind:ref {tabID} {tabName} {...restProps}>
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>
				<div class="flex w-full items-center justify-between">
					Connections
					<Button size="sm" variant="ghost" onclick={() => refresh()}>
						<RefreshCw />
					</Button>
				</div>
			</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each Array.from(postgresConnectionsMap.entries()) as [key, connection]}
						<Sidebar.MenuItem
							class="{getColorClass(connection.Colour)} bg-opacity-20 hover:bg-opacity-25"
						>
							<Collapsible.Root>
								<Collapsible.Trigger onclick={() => establishConnection(connection.ID)}>
									<Sidebar.MenuButton class="transform-none select-none transition-none">
										<ChevronRight className="transition-transform" />

										{connection.Name}
									</Sidebar.MenuButton>
								</Collapsible.Trigger>
								<Collapsible.Content>
									<Sidebar.MenuSub>
										{#if loadingMap.get(connection.ID)}
											<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
											<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
											<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
											<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
										{:else if (connectionDatabasesMap.get(connection.ID) || []).length > 0}
											{#each connectionDatabasesMap.get(connection.ID) || [] as databaseID}
												<Collapsible.Root open={databasesMap.get(databaseID)?.IsActive}>
													<Collapsible.Trigger
														onclick={() => establishDatabaseConnection(connection.ID, databaseID)}
													>
														<ContextMenu.Root>
															<ContextMenu.Trigger>
																<Sidebar.MenuButton
																	class="transform-none select-none transition-none"
																	aria-disabled={!databasesMap.get(databaseID)?.IsActive}
																>
																	<ChevronRight className="transition-transform" />
																	{databasesMap.get(databaseID)?.Name}
																	{#if databasesMap.get(databaseID)?.IsActive}
																		<Activity color="#4fff4d" />
																	{/if}
																</Sidebar.MenuButton>
															</ContextMenu.Trigger>
															<ContextMenu.Content>
																<ContextMenu.Item onclick={() => terminateDBConnection(databaseID)}
																	>Disconnect
																</ContextMenu.Item>
																<ContextMenu.Item onclick={() => refreshDB(databaseID)}
																	>Refresh
																</ContextMenu.Item>
															</ContextMenu.Content>
														</ContextMenu.Root>
													</Collapsible.Trigger>
													<Collapsible.Content>
														<Sidebar.MenuSub>
															{#if dbLoadingMap.get(databaseID)}
																<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
																<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
																<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
																<Skeleton class="my-1 h-[20px] w-[100px] rounded-full" />
															{:else if (databasesMap.get(databaseID)?.Tables || []).length > 0}
																{#each databasesMap.get(databaseID)?.Tables || [] as table}
																	<Sidebar.MenuButton>
																		<Table2 color="#fd6868" strokeWidth={2} size={25} />
																		<p>{table}</p>
																	</Sidebar.MenuButton>
																{/each}
															{:else}
																No tables found
															{/if}
														</Sidebar.MenuSub>
													</Collapsible.Content>
												</Collapsible.Root>
											{/each}
										{:else}
											No databases found
										{/if}
									</Sidebar.MenuSub>
								</Collapsible.Content>
							</Collapsible.Root>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<div class="flex flex-row items-center">
			<Plus size={25} class="align-left mx-2" />
			<SettingsDialog />
			<a href="/settings" class="ml-auto"><Settings size={25} color="#f05c5c" href="/settings" /></a
			>
		</div>
		<!-- <Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton>
					
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu> -->
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
