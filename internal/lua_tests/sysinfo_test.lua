function main()
	 initSystem, err = api.sys.getInitSystem()
	 if err ~= nil then return false end
	 distro = api.sys.getDistro()
	
	 return initSystem ~= "" and distro.Name ~= ""
end
