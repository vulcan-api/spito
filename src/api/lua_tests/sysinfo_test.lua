function main()
	 initSystem, err = api.sys.GetInitSystem()
	 if err ~= nil then return false end
	 distro = api.sys.GetDistro()
	
	 return initSystem ~= "" and distro.Name ~= ""
end
