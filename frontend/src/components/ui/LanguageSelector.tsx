import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { GlobeAltIcon } from '@heroicons/react/24/outline';
import { ChevronDownIcon, CheckCircleIcon } from '../icons';

const LanguageSelector: React.FC = () => {
  const { i18n, t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const languages = [
    { code: 'en', name: 'English', shortName: 'EN', region: 'United States' },
    { code: 'de', name: 'Deutsch', shortName: 'DE', region: 'Germany' },
  ];

  const getValidLanguage = (lang: string): string => {
    if (!lang) return 'en';
    const cleanLang = lang.toLowerCase().split('-')[0];
    return ['en', 'de'].includes(cleanLang) ? cleanLang : 'en';
  };

  const currentLangCode = getValidLanguage(i18n.language);
  const currentLanguage = languages.find(lang => lang.code === currentLangCode) || languages[0];

  useEffect(() => {
    if (i18n.language && !['en', 'de'].includes(i18n.language.split('-')[0])) {
      i18n.changeLanguage('en');
      localStorage.setItem('i18nextLng', 'en');
    }
  }, [i18n]);

  const handleLanguageChange = (languageCode: string) => {
    i18n.changeLanguage(languageCode);
    localStorage.setItem('i18nextLng', languageCode);
    setIsOpen(false);
  };

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={toggleDropdown}
        className="flex items-center space-x-2 px-3 py-2 text-sm text-gray-700 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 rounded-md"
        aria-label={t('dashboard.language')}
        aria-expanded={isOpen}
      >
        <GlobeAltIcon className="h-5 w-5" />
        <span className="text-xs bg-blue-100 px-2 py-1 rounded font-bold text-blue-800 border border-blue-200">
          {currentLanguage.shortName}
        </span>
        <span className="hidden md:block font-medium">{currentLanguage.name}</span>
        <ChevronDownIcon className={`h-4 w-4 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-56 bg-white rounded-md shadow-lg border border-gray-200 z-50">
          <div className="py-1" role="menu">
            {languages.map((language) => (
              <button
                key={language.code}
                onClick={() => handleLanguageChange(language.code)}
                className={`flex items-center w-full px-4 py-3 text-sm hover:bg-gray-100 transition-colors ${
                  currentLangCode === language.code
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-gray-700'
                }`}
                role="menuitem"
              >
                <div className="flex items-center justify-between w-full">
                  <div className="flex items-center">
                    <span className="mr-3 text-xs bg-gray-100 px-2 py-1 rounded font-mono font-semibold">{language.shortName}</span>
                    <div className="flex flex-col items-start">
                      <span className="font-medium">{language.name}</span>
                      <span className="text-xs text-gray-500">{language.region}</span>
                    </div>
                  </div>
                  {currentLangCode === language.code && (
                    <CheckCircleIcon className="h-4 w-4 text-blue-600" />
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default LanguageSelector; 
