docker_compose("./docker-compose.yaml", '.env', "virtual_court_data_migration", ['migrate'])

def is_nomigrate(profile):
    return profile.lower() == 'nomigrate'

for service_name in range(1, 6):
    profile_key = f'CHUNK_FOLDER_{service_name}_PROFILE'
    condition_name = f'service_{service_name}_condition'

    def condition(ctx):
        nomigrate_profiles = ['nomigrate', '']
        return is_nomigrate(ctx.env.get(profile_key, '')) not in nomigrate_profiles

    globals()[condition_name] = condition