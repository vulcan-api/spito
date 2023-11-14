function main()
	 initSystem = GetInitSystem()
	 distro = GetDistro()
	 return initSystem ~= "" and distro.Version~= "" and distro.Name ~= ""
end
