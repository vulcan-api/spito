-- Command:
-- "BaderBc/my-configs-spito" is ruleset and implementationset
	--spito pass rule BaderBC/my-configs-spito astro-nvim


-- (Optionally) check if package database is updated):
function newest_package_database()
	local database_rule = rule_import("BaderBC/spito-pacman", "newest-databases", {
		pass_required = false,  -- If true, even in case when this (not imported) rule returns true, status won't be PASSED, default true
		auto_implement = false, -- If false, rule checker cannot automatically suggest user any implementation which will make rule pass, default true
		root_func = "main",  		-- Default main
	})

	local database_rule_result = database_rule() -- If main rule function takes some arguments we can pass them here
	if database_rule_result then return true end
	
	-- This will be suggestion for user - how to make this rule pass
	local database_impl = implementation_import("BaderBC/spito-pacman", "database-update", {
		optional = true, -- Info for user that this is optional and there won't be any consequences for skipping it
	})
	database_impl() -- If main rule function takes some arguments we can pass them here

	return false
end

-- check if neovim is installed:
function neovim()
	local does_neovim_exists = api.pkg.Exists("neovim")
	if not does_neovim_exists then
		neovim_install_impl = implementation_import("BaderBC/spito-packages", "neovim"), {
			root_func = "install", -- default main
		})
		neovim_install_impl()
		return false
	end
	return true
end

-- check if astro-nvim is installed
function astro_nvim()
	local isAstroInstalled = api.fs.PathExists("~/.config/nvim/lua/astronvim")
	if not isAstroInstalled then
		local astro_impl = implementation_import("BaderBC/my-configs-spito", "astro-nvim")
	end
	return isAstroInstalled
end

-- main function
function main()
	newest_package_database()

	return neovim() and astro_nvim()
end
