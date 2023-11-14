function main()
	 initSystem, err = GetInitSystem()
	 if err ~= nil then return false end
	 distro = GetDistro()
	 return initSystem ~= "" and distro.Name ~= ""
end
