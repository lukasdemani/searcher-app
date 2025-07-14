import React from 'react';
import { useTranslation } from 'react-i18next';
import { LanguageIcon } from '@heroicons/react/24/outline';
import { ChevronDownIcon, CheckCircleIcon } from '../icons';

const LanguageSelector: React.FC = () => {
  const { i18n, t } = useTranslation();

  const languages = [
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
  ];

  const currentLanguage = languages.find(lang => lang.code === i18n.language) || languages[0];

  const handleLanguageChange = (languageCode: string) => {
    i18n.changeLanguage(languageCode);
  };

  return (
    <div className="relative group">
      <button
        type="button"
        className="flex items-center space-x-2 px-3 py-2 text-sm text-gray-700 hover:text-gray-900 focus:outline-none"
        aria-label={t('dashboard.language')}
      >
        <LanguageIcon className="h-5 w-5" />
        <span className="hidden sm:block">{currentLanguage.flag}</span>
        <span className="hidden md:block">{currentLanguage.name}</span>
        <ChevronDownIcon className="h-4 w-4" />
      </button>

      <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
        <div className="py-1" role="menu">
          {languages.map((language) => (
            <button
              key={language.code}
              onClick={() => handleLanguageChange(language.code)}
              className={`flex items-center w-full px-4 py-2 text-sm hover:bg-gray-100 transition-colors ${
                i18n.language === language.code
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-gray-700'
              }`}
              role="menuitem"
            >
              <span className="mr-3 text-lg">{language.flag}</span>
              <span>{language.name}</span>
              {i18n.language === language.code && (
                <CheckCircleIcon className="ml-auto h-4 w-4 text-blue-600" />
              )}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
};

export default LanguageSelector; 